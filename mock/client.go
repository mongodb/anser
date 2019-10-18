package mock

import (
	"context"

	"github.com/mongodb/anser/client"
	"github.com/mongodb/ftdc/bsonx"
)

type Client struct {
	Databases map[string]*Database
}

func NewClient() *Client {
	return &Client{
		Databases: map[string]*Database{},
	}
}

func (c *Client) Connect(ctx context.Context) error    { return nil }
func (c *Client) Disconnect(ctx context.Context) error { return nil }
func (c *Client) Database(name string) client.Database {
	if db, ok := c.Databases[name]; ok {
		return db
	}

	c.Databases[name] = &Database{DBName: name, Collections: map[string]*Collection{}}
	return c.Databases[name]
}

func (c *Client) ListDatabaseNames(ctx context.Context, query interface{}) ([]string, error) {
	names := make([]string, 0, len(c.Databases))
	for key := range c.Databases {
		names = append(names, key)
	}
	return names, nil
}

type Database struct {
	DBName      string
	Collections map[string]*Collection
}

func (d *Database) Name() string          { return d.DBName }
func (d *Database) Client() client.Client { return nil }
func (d *Database) Collection(name string) client.Collection {
	if coll, ok := d.Collections[name]; ok {
		return coll
	}

	d.Collections[name] = &Collection{CollName: name, SingleResult: NewSingleResult()}
	return d.Collections[name]
}
func (d *Database) RunCommand(ctx context.Context, cmd interface{}) client.SingleResult { return nil }
func (d *Database) RunCommandCursor(ctx context.Context, cmd interface{}) (client.Cursor, error) {
	return nil, nil
}

type Collection struct {
	CollName         string
	UpdateResult     client.UpdateResult
	SingleResult     *SingleResult
	InsertManyResult client.InsertManyResult
	InsertOneResult  client.InsertOneResult
}

func (c *Collection) Name() string { return c.CollName }
func (c *Collection) Aggregate(ctx context.Context, pipe interface{}) (client.Cursor, error) {
	return nil, nil
}
func (c *Collection) Find(ctx context.Context, query interface{}) (client.Cursor, error) {
	return nil, nil
}
func (c *Collection) FindOne(ctx context.Context, query interface{}) client.SingleResult {
	return c.SingleResult
}
func (c *Collection) InsertOne(ctx context.Context, doc interface{}) (*client.InsertOneResult, error) {
	return &c.InsertOneResult, nil
}
func (c *Collection) InsertMany(ctx context.Context, docs []interface{}) (*client.InsertManyResult, error) {
	return &c.InsertManyResult, nil
}

func (c *Collection) ReplaceOne(ctx context.Context, query, update interface{}) (*client.UpdateResult, error) {
	return &c.UpdateResult, nil
}
func (c *Collection) UpdateOne(ctx context.Context, query, update interface{}) (*client.UpdateResult, error) {
	return &c.UpdateResult, nil
}
func (c *Collection) UpdateMany(ctx context.Context, query, update interface{}) (*client.UpdateResult, error) {
	return &c.UpdateResult, nil
}

type Cursor struct {
	CurrentValue   []byte
	AllError       error
	CloseError     error
	DecodeError    error
	ErrError       error
	CursorID       int64
	NextCallsCount int
	MaxNextCalls   int
}

func (c *Cursor) Current() []byte                               { return c.CurrentValue }
func (c *Cursor) All(ctx context.Context, in interface{}) error { return c.AllError }
func (c *Cursor) Close(ctx context.Context) error               { return c.CloseError }
func (c *Cursor) Decode(in interface{}) error                   { return c.DecodeError }
func (c *Cursor) Err() error                                    { return c.ErrError }
func (c *Cursor) ID() int64                                     { return c.CursorID }
func (c *Cursor) Next(ctx context.Context) bool {
	c.NextCallsCount++

	return c.MaxNextCalls > c.NextCallsCount
}

type SingleResult struct {
	DecodeError      error
	DecodeBytesError error
	DecodeBytesValue []byte
	ErrorValue       error
}

func NewSingleResult() *SingleResult {
	doc := bsonx.NewDocument()
	val, _ := doc.MarshalBSON()

	return &SingleResult{DecodeBytesValue: val}
}

func (sr *SingleResult) Decode(in interface{}) error  { return sr.DecodeError }
func (sr *SingleResult) DecodeBytes() ([]byte, error) { return sr.DecodeBytesValue, sr.DecodeBytesError }
func (sr *SingleResult) Err() error                   { return sr.ErrorValue }
