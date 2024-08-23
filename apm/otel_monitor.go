// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/sometimes"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTracerName          = "go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	responseBytesAttribute     = "db.response_bytes"
	strippedStatementAttribute = "db.statement.stripped"
)

// config is used to configure the mongo tracer.
type config struct {
	TracerProvider trace.TracerProvider

	Tracer trace.Tracer

	CommandAttributeDisabled bool

	CommandTransformerFunc CommandTransformer
}

// newConfig returns a config with all Options set.
func newConfig(opts ...Option) config {
	cfg := config{
		TracerProvider:           otel.GetTracerProvider(),
		CommandTransformerFunc:   transformCommand,
		CommandAttributeDisabled: true,
	}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	cfg.Tracer = cfg.TracerProvider.Tracer(
		defaultTracerName,
	)
	return cfg
}

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(cfg *config) {
		if provider != nil {
			cfg.TracerProvider = provider
		}
	})
}

// WithCommandAttributeDisabled specifies if the MongoDB command is added as an attribute to Spans or not.
// This is disabled by default and the MongoDB command will not be added as an attribute
// to Spans if this option is not provided.
func WithCommandAttributeDisabled(disabled bool) Option {
	return optionFunc(func(cfg *config) {
		cfg.CommandAttributeDisabled = disabled
	})
}

// CommandTransformer defines a function that transforms a MongoDB command attribute.
// If the function returns an empty string, the attribute will not be added to the Span.
type CommandTransformer func(command bson.Raw) bson.Raw

// WithCommandAttributeTransformer specifies a function to transform the MongoDB command attribute.
func WithCommandAttributeTransformer(transformer CommandTransformer) Option {
	return optionFunc(func(cfg *config) {
		if transformer != nil {
			cfg.CommandTransformerFunc = transformer
		} else {
			cfg.CommandTransformerFunc = transformCommand
		}
	})
}

type spanKey struct {
	ConnectionID string
	RequestID    int64
}

type monitor struct {
	sync.Mutex
	spans map[spanKey]trace.Span
	cfg   config
}

func (m *monitor) Started(ctx context.Context, evt *event.CommandStartedEvent) {
	var spanName string

	hostname, port := peerInfo(evt)

	attrs := []attribute.KeyValue{
		semconv.DBSystemMongoDB,
		semconv.DBOperation(evt.CommandName),
		semconv.DBName(evt.DatabaseName),
		semconv.NetPeerName(hostname),
		semconv.NetPeerPort(port),
		semconv.NetTransportTCP,
	}
	if !m.cfg.CommandAttributeDisabled {
		statementAttributes, err := m.dbStatementAttributes(evt)
		if err == nil {
			attrs = append(attrs, statementAttributes...)
		} else {
			grip.ErrorWhen(sometimes.Percent(10), errors.Wrap(err, "getting command attributes"))
		}
	}
	if collection, err := extractCollection(evt); err == nil && collection != "" {
		spanName = collection + "."
		attrs = append(attrs, semconv.DBMongoDBCollection(collection))
	}
	spanName += evt.CommandName
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	}
	_, span := m.cfg.Tracer.Start(ctx, spanName, opts...)
	key := spanKey{
		ConnectionID: evt.ConnectionID,
		RequestID:    evt.RequestID,
	}
	m.Lock()
	m.spans[key] = span
	m.Unlock()
}

func (m *monitor) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	span, ok := m.getSpan(&evt.CommandFinishedEvent)
	if !ok {
		return
	}
	span.SetAttributes(attribute.Int(responseBytesAttribute, len(evt.Reply)))
	span.End()
}

func (m *monitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	span, ok := m.getSpan(&evt.CommandFinishedEvent)
	if !ok {
		return
	}
	span.SetStatus(codes.Error, evt.Failure.Error())
	span.End()
}

func (m *monitor) getSpan(evt *event.CommandFinishedEvent) (trace.Span, bool) {
	key := spanKey{
		ConnectionID: evt.ConnectionID,
		RequestID:    evt.RequestID,
	}
	m.Lock()
	span, ok := m.spans[key]
	if ok {
		delete(m.spans, key)
	}
	m.Unlock()

	return span, ok
}

func (m *monitor) dbStatementAttributes(evt *event.CommandStartedEvent) ([]attribute.KeyValue, error) {
	var attributes []attribute.KeyValue
	command := m.cfg.CommandTransformerFunc(evt.Command)
	// The transformer function can return nil if it doesn't want the contents of the statement
	// included as an attribute. An example would be if it contains sensitive data.
	if command == nil {
		return nil, nil
	}

	section, err := operationSection(evt.CommandName, command)
	if err != nil {
		return nil, errors.Wrap(err, "extracting statement")
	}

	formattedStmt, err := formatStatement(section, false)
	if err != nil {
		return nil, errors.Wrap(err, "formatting statement")
	}
	if formattedStmt != "" {
		// Add an attribute with the actual command. This is useful for reproducing queries
		// to debug why they're taking longer than expected.
		attributes = append(attributes, semconv.DBStatement(formattedStmt))
	}

	strippedStatement, err := formatStatement(section, true)
	if err != nil {
		return nil, errors.Wrap(err, "formatting stripped statement")
	}
	if strippedStatement != "" {
		// Add an attribute with the command stripped of its values. This is useful for grouping
		// all the instances of a query together even when their values differ.
		attributes = append(attributes, attribute.String(strippedStatementAttribute, strippedStatement))
	}

	return attributes, nil
}

// TODO sanitize values where possible, then reenable `db.statement` span attributes default.
// TODO limit maximum size.
func transformCommand(command bson.Raw) bson.Raw {
	return command
}

// extractCollection extracts the collection for the given mongodb command event.
// For CRUD operations, this is the first key/value string pair in the bson
// document where key == "<operation>" (e.g. key == "insert").
// For database meta-level operations, such a key may not exist.
func extractCollection(evt *event.CommandStartedEvent) (string, error) {
	elt, err := evt.Command.IndexErr(0)
	if err != nil {
		return "", err
	}
	if key, err := elt.KeyErr(); err == nil && key == evt.CommandName {
		var v bson.RawValue
		if v, err = elt.ValueErr(); err != nil || v.Type != bson.TypeString {
			return "", err
		}
		return v.StringValue(), nil
	}
	return "", fmt.Errorf("collection name not found")
}

// NewMonitor creates a new mongodb event CommandMonitor.
func NewMonitor(opts ...Option) *event.CommandMonitor {
	cfg := newConfig(opts...)
	m := &monitor{
		spans: make(map[spanKey]trace.Span),
		cfg:   cfg,
	}
	return &event.CommandMonitor{
		Started:   m.Started,
		Succeeded: m.Succeeded,
		Failed:    m.Failed,
	}
}

func peerInfo(evt *event.CommandStartedEvent) (hostname string, port int) {
	hostname = evt.ConnectionID
	port = 27017
	if idx := strings.IndexByte(hostname, '['); idx >= 0 {
		hostname = hostname[:idx]
	}
	if idx := strings.IndexByte(hostname, ':'); idx >= 0 {
		port = func(p int, e error) int { return p }(strconv.Atoi(hostname[idx+1:]))
		hostname = hostname[:idx]
	}
	return hostname, port
}

func formatStatement(statement bson.Raw, stripped bool) (string, error) {
	var err error
	if stripped {
		statement, err = stripDocument(statement)
		if err != nil {
			return "", errors.Wrap(err, "stripping section values")
		}
	}

	b, err := bson.MarshalExtJSON(statement, false, false)
	if err != nil {
		return "", errors.Wrap(err, "marshalling to extended JSON")
	}

	var buf bytes.Buffer
	err = json.Indent(&buf, b, "", "  ")
	return buf.String(), errors.Wrap(err, "indenting JSON")
}

func operationSection(commandName string, raw bson.Raw) (bson.Raw, error) {
	switch commandName {
	case "aggregate":
		return extractAggregation(raw)
	case "delete":
		return extractDelete(raw)
	case "find":
		return extractFind(raw)
	case "findAndModify":
		return extractFindAndModify(raw)
	case "update":
		return extractUpdate(raw)
	case "insert":
		return extractInsert(raw)
	default:
		return raw, nil
	}
}

func extractAggregation(statement bson.Raw) (bson.Raw, error) {
	elems, err := statement.Elements()
	if err != nil {
		return nil, errors.Wrap(err, "getting elements for aggregation statement")
	}

	var pipelineDoc bson.D
	for _, elem := range elems {
		if elem.Key() == "pipeline" {
			pipelineDoc = bson.D{bson.E{Key: elem.Key(), Value: elem.Value()}}
			break
		}
	}

	return bson.Marshal(pipelineDoc)
}

func extractDelete(statement bson.Raw) (bson.Raw, error) {
	elems, err := statement.Elements()
	if err != nil {
		return nil, errors.Wrap(err, "getting elements for delete statement")
	}

	for _, elem := range elems {
		if elem.Key() == "deletes" {
			deletesArray, ok := elem.Value().ArrayOK()
			if !ok {
				break
			}
			vals, err := deletesArray.Values()
			if err != nil {
				return nil, errors.Wrap(err, "getting values for deletes array")
			}
			if len(vals) == 0 {
				break
			}
			return vals[0].Value, nil
		}
	}
	return nil, nil
}

func extractFind(statement bson.Raw) (bson.Raw, error) {
	elems, err := statement.Elements()
	if err != nil {
		return nil, errors.Wrap(err, "getting elements for find statement")
	}

	findFields := []string{"filter", "sort", "limit", "hint"}
	var findDoc bson.D
	for _, elem := range elems {
		if utility.StringSliceContains(findFields, elem.Key()) {
			findDoc = append(findDoc, bson.E{Key: elem.Key(), Value: elem.Value()})
		}
	}

	return bson.Marshal(findDoc)
}

func extractFindAndModify(statement bson.Raw) (bson.Raw, error) {
	elems, err := statement.Elements()
	if err != nil {
		return nil, errors.Wrap(err, "getting elements for findAndModify statement")
	}

	findFields := []string{"query", "update", "upsert"}
	var findDoc bson.D
	for _, elem := range elems {
		if utility.StringSliceContains(findFields, elem.Key()) {
			findDoc = append(findDoc, bson.E{Key: elem.Key(), Value: elem.Value()})
		}
	}

	return bson.Marshal(findDoc)
}

func extractUpdate(statement bson.Raw) (bson.Raw, error) {
	elems, err := statement.Elements()
	if err != nil {
		return nil, errors.Wrap(err, "getting elements for update statement")
	}

	for _, elem := range elems {
		if elem.Key() == "updates" {
			updatesArray, ok := elem.Value().ArrayOK()
			if !ok {
				break
			}
			vals, err := updatesArray.Values()
			if err != nil {
				return nil, errors.Wrap(err, "getting values for updates array")
			}
			if len(vals) == 0 {
				break
			}
			return vals[0].Value, nil
		}
	}
	return nil, nil
}

func extractInsert(statement bson.Raw) (bson.Raw, error) {
	elems, err := statement.Elements()
	if err != nil {
		return nil, errors.Wrap(err, "getting elements for insert statement")
	}

	insertFields := []string{"ordered", "documents"}
	var insertDoc bson.D
	for _, elem := range elems {
		if utility.StringSliceContains(insertFields, elem.Key()) {
			insertDoc = append(insertDoc, bson.E{Key: elem.Key(), Value: elem.Value()})
		}
	}

	return bson.Marshal(insertDoc)
}

func stripDocument(doc bson.Raw) (bson.Raw, error) {
	elems, err := doc.Elements()
	if err != nil {
		return nil, errors.Wrap(err, "enumerating document elements")
	}
	strippedDocument := bson.D{}
	for _, elem := range elems {
		elemValue, err := stripValue(elem.Value())
		if err != nil {
			return nil, errors.Wrap(err, "stripping document values")
		}
		strippedDocument = append(strippedDocument, bson.E{Key: elem.Key(), Value: elemValue})
	}

	return bson.Marshal(strippedDocument)
}

func stripValue(val bson.RawValue) (bson.RawValue, error) {
	switch elemType := val.Type.String(); elemType {
	case "embedded document":
		strippedSubdocument, err := stripDocument(val.Document())
		return bson.RawValue{Type: bson.TypeEmbeddedDocument, Value: strippedSubdocument}, errors.Wrap(err, "stripping subdocument")
	case "array":
		values, err := val.Array().Values()
		if err != nil {
			return bson.RawValue{}, errors.Wrap(err, "getting array values")
		}
		arr := bson.A{}
		for _, val := range values {
			strippedVal, err := stripValue(val)
			if err != nil {
				return bson.RawValue{}, errors.Wrap(err, "stripping values for array member")
			}
			arr = append(arr, strippedVal)
		}
		arr = compactArray(arr)
		_, encodedArray, err := bson.MarshalValue(arr)
		return bson.RawValue{Type: bson.TypeArray, Value: encodedArray}, errors.Wrap(err, "encoding array")
	default:
		_, encodedValue, err := bson.MarshalValue(fmt.Sprintf("<%s>", val.Type.String()))
		return bson.RawValue{Type: bson.TypeString, Value: encodedValue}, errors.Wrap(err, "encoding value")
	}
}

func compactArray(arr bson.A) bson.A {
	compactedArray := make(bson.A, 0, len(arr))
	types := make(map[string]bool)
	for _, elem := range arr {
		elemVal, ok := elem.(bson.RawValue)
		if !ok {
			return arr
		}
		if elemVal.Type != bson.TypeString {
			return arr
		}
		valString, ok := elemVal.StringValueOK()
		if !ok {
			return arr
		}
		if !types[valString] {
			compactedArray = append(compactedArray, elem)
		}
		types[valString] = true
	}

	return compactedArray
}
