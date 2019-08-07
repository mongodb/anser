package client

import (
	"github.com/mongodb/anser/model"
	"github.com/mongodb/ftdc/bsonx"
)

// Processor defines the process for processing a stream of
// documents using a Cursor.
type Processor interface {
	Load(Client, model.Namespace, map[string]interface{}) Cursor
	Migrate(Cursor) error
}

// MigrationOperation defines the function object that performs
// the transformation in the manual migration migrations. Register
// these functions using RegisterMigrationOperation.
//
// Implementors of MigrationOperations are responsible for
// implementing idempotent operations.
type MigrationOperation func(Client, *bsonx.Document) error
