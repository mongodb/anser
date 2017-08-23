/*
Anser Migrations

The anser package defines supports three differnt migration definition
forms: the SimpleMigration for single document updates, using
MongoDB's update query form, the ManualMigration which takes as input
a single document and provides an MongoDB session operation for manual
migration, and the AggregateProducer

Migrations themselves are executed as amboy.Jobs either serially or in
an amboy Queue. Although round-trippable serialization is not a strict
requirement of running migrations as amboy Jobs, these migrations
support round-trippable BSON serialization and thus distributed
queues.

SimpleMigration

Use simple migrations to rename a field in a document or change the
structure of a document using MongoDB queries. Prefer these operations
to running actual queries for the rate-limiting properties of the
Anser executor.

ManualMigration

Use manual migrations when you need to perform a migration operation
that requires application logic, results in the creation of new
documents, or requires destructive modification of the source
document.

AggregateProducer

Use aggregate producer for processing using application logic, an
iterator of documents. This is similar to the manual migration but
allows reduce-like operations, or even destructive operations.
*/

package anser

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Simple Migration defines a single-document operation, performing a
// single document update that operates on one collection.
type SimpleMigration struct {
	// ID returns the _id field of the document that is the
	// subject of the migration.
	ID interface{} `bson:"id" json:"id" yaml:"id"`

	// Update is a specification for the update
	// operation. This is converted to a bson.M{} (but is typed as
	// a map to allow migration implementations to *not* depend upon
	// bson.M
	Update map[string]interface{} `bson:"update" json:"update" yaml:"update"`

	// Migration holds the ID of the migration operation,
	// typically the name of the class of migration and
	Migration string `bson:"migration_id" json:"migration_id" yaml:"migration_id"`

	// Namespace holds a struct that describes which database and
	// collection where the migration should run
	Namespace Namespace `bson:"namespace" json:"namespace" yaml:"namespace"`
}

// ManualMigrationOperation defines the function object that performs
// the transformation in the manual migration migrations. Register
// these functions using RegisterManualMigrationOperation.
//
// Implementors of ManualMigrationOperations are responsible for
// implementating idempotent operations.
type ManualMigrationOperation func(*mgo.Session, bson.Raw) error

// ManualMigrations defines an operations that runs an arbitrary
// function given the input of an mgo.Session pointer and
type ManualMigration struct {
	// ID returns the _id field of the document that is the
	// input of the migration.
	ID interface{} `bson:"id" json:"id" yaml:"id"`

	// The name of the migration function to run. This function
	// must be defined in the migration registry
	OperationName string `bson:"op_name" json:"op_name" yaml:"op_name"`

	// Migration holds the ID of the migration operation,
	// typically the name of the class of migration and
	Migration string `bson:"migration_id" json:"migration_id" yaml:"migration_id"`

	// Namespace holds a struct that describes which database and
	// collection where the query for the input document should run.
	Namespace Namespace `bson:"namespace" json:"namespace" yaml:"namespace"`
}

// AggregateProducer is a migration definition form that has, that can
// processes a stream of documents, using an implementation of the
// DocumentProducer interface.
type AggregateProducer struct {
	Query map[string]interface{} `bson:"query" json:"query" yaml:"query"`

	// The name of a registered DocumentProcessor implementation.
	// Because the producer isn't versioned, changes to the
	// implementation of the Producer between migration runs, may
	// complicate idempotency requirements.
	ProcessorName string `bson:"producer" json:"producer" yaml:"producer"`

	// Migration holds the ID of the migration operation,
	// typically the name of the class of migration and
	Migration string `bson:"migration_id" json:"migration_id" yaml:"migration_id"`

	// Namespace holds a struct that describes which database and
	// collection where the query for the input document should run.
	Namespace Namespace `bson:"namespace" json:"namespace" yaml:"namespace"`
}
