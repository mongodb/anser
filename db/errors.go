package db

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrNotFound = errors.New("document not found")

func ResultsNotFound(err error) bool {
	return errors.Cause(err) == ErrNotFound || errors.Cause(err) == mongo.ErrNoDocuments
}
