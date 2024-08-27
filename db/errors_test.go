package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestResultsPredicate(t *testing.T) {
	assert := assert.New(t)

	assert.False(ResultsNotFound(errors.New("foo")))
	assert.False(ResultsNotFound(nil))
	assert.False(ResultsNotFound(errors.New("not found")))
	assert.True(ResultsNotFound(mongo.ErrNoDocuments))
	assert.True(ResultsNotFound(ErrNotFound))
}
