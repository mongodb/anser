/*
DocumentProcessor

The DocuumentProcessor is an interface that you can implement for
migrations to process groups of documents. Rather than defining
migrations that operate on a single document, these migrations have
access to an iterator and operate on many documents.

*/
package anser

import (
	"github.com/tychoish/anser/model"
	mgo "gopkg.in/mgo.v2"
)

// DocumentProcessor defines the process for processing a stream of
// documents using a DocumentIterator, which resembles mgo's Iter
// operation.
type DocumentProcessor interface {
	Load(model.Namespace, map[string]interface{}) DocumentIterator
	Migrate(DocumentIterator) error
}

// DocumentIterator is a more narrow subset of mgo's Iter type that
// provides the opportunity to mock results, and avoids a strict
// dependency between mgo and migrations definitions.
type DocumentIterator interface {
	Next(interface{}) bool
	Close() error
	Err() error
}

// NewCombinedIterator produces a DocumentIterator that is an
// mgo.Iter, with a modified Close() method that also closes the
// provided mgo session after closing the iterator.
func NewCombinedIterator(ses *mgo.Session, iter *mgo.Iter) DocumentIterator {
	return combinedCloser{
		Iter: iter,
		ses:  ses,
	}
}

type combinedCloser struct {
	*mgo.Iter
	ses *mgo.Session
}

func (c combinedCloser) Close() error {
	err := c.Iter.Close()
	c.ses.Close()
	return err
}
