package client

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Client interface {
	Connect(context.Context) error
	Disconnect(context.Context) error

	Database(string) Database
	ListDatabaseNames(context.Context, interface{}) ([]string, error)
}

type Database interface {
	Client() Client
	Collection(string) Collection
	Name() string
	RunCommand(context.Context, interface{}) SingleResult
	RunCommandCursor(context.Context, interface{}) (Cursor, error)
}

type Collection interface {
	Aggregate(context.Context, interface{}, ...options.Lister[options.AggregateOptions]) (Cursor, error)
	Find(context.Context, interface{}, ...options.Lister[options.FindOptions]) (Cursor, error)
	FindOne(context.Context, interface{}, ...options.Lister[options.FindOneOptions]) SingleResult
	Name() string
	ReplaceOne(context.Context, interface{}, interface{}, ...options.Lister[options.ReplaceOptions]) (*UpdateResult, error)
	UpdateOne(context.Context, interface{}, interface{}, ...options.Lister[options.UpdateOneOptions]) (*UpdateResult, error)
	UpdateMany(context.Context, interface{}, interface{}, ...options.Lister[options.UpdateManyOptions]) (*UpdateResult, error)
	InsertOne(context.Context, interface{}) (*InsertOneResult, error)
	InsertMany(context.Context, []interface{}) (*InsertManyResult, error)
}

type SingleResult interface {
	Decode(interface{}) error
	Raw() ([]byte, error)
	Err() error
}

type Cursor interface {
	Current() []byte
	All(context.Context, interface{}) error
	Close(context.Context) error
	Decode(interface{}) error
	Err() error
	ID() int64
	Next(context.Context) bool
}

type InsertOneResult = mongo.InsertOneResult
type InsertManyResult = mongo.InsertManyResult
type UpdateResult = mongo.UpdateResult
