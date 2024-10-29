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
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/event"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTracerName      = "go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	responseBytesAttribute = "db.response_bytes"
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
type CommandTransformer func(command bson.Raw) string

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
		if stmt := m.cfg.CommandTransformerFunc(evt.Command); stmt != "" {
			attrs = append(attrs, semconv.DBStatement(stmt))
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
	span.SetAttributes(attribute.Int(response_bytes_attribute, len(evt.Reply)))
	span.End()
}

func (m *monitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	span, ok := m.getSpan(&evt.CommandFinishedEvent)
	if !ok {
		return
	}
	span.SetStatus(codes.Error, evt.Failure)
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

// TODO sanitize values where possible, then reenable `db.statement` span attributes default.
// TODO limit maximum size.
func transformCommand(command bson.Raw) string {
	b, _ := bson.MarshalExtJSON(command, false, false)
	return string(b)
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
		if v, err = elt.ValueErr(); err != nil || v.Type != bsontype.String {
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
