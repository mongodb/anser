/*
DocumentProcessor

The DocuumentProcessor is an interface that you can implement for
migrations to process groups of documents. Rather than defining
migrations that operate on a single document, these migrations have
access to an iterator and operate on many documents.

The document processor system wraps the MGO driver internals using
interfaces provided by the anser/db package.

*/
package anser

import (
	"github.com/tychoish/anser/db"
	"github.com/tychoish/anser/model"
)

// DocumentProcessor defines the process for processing a stream of
// documents using a DocumentIterator, which resembles mgo's Iter
// operation.
type DocumentProcessor interface {
	Load(model.Namespace, map[string]interface{}) db.Iterator
	Migrate(db.Iterator) error
}
