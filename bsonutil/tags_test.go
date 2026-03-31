package bsonutil

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTag(t *testing.T) {
	assert := assert.New(t)

	t.Run("Invalid", func(t *testing.T) {
		var FieldOne string = `bson:"tag1"`
		_, err := Tag(FieldOne, "")
		assert.Error(err)
	})

	t.Run("MissingStruct", func(t *testing.T) {
		type sOne struct {
		}
		_, err := Tag(sOne{}, "FieldOne")
		assert.Error(err)
	})
	
	t.Run("MissingTag", func(t *testing.T) {
		type sTwo struct {
			FieldOne string
		}
		tagVal, err := Tag(sTwo{}, "FieldOne")
		assert.NoError(err)
		assert.Equal(tagVal, "")
	})

	t.Run("FetchTag", func(t *testing.T) {
		type sThree struct {
			FieldOne string `bson:"tag1"`
		}
		tagVal, err := Tag(sThree{}, "FieldOne")
		assert.NoError(err)
		assert.Equal(tagVal, "tag1")
	})

	t.Run("Slice", func(t *testing.T) {
		type sThree struct {
			FieldOne string `bson:"tag1"`
		}
		tagVal, err := Tag([]sThree{}, "FieldOne")
		assert.NoError(err)
		assert.Equal(tagVal, "tag1")
	})

	t.Run("IgnoreModifiers", func(t *testing.T){
		type sFour struct {
			FieldOne string `bson:"tag1,omitempty"`
		}
		tagVal, err := Tag(sFour{}, "FieldOne")
		assert.NoError(err)
		assert.Equal(tagVal, "tag1")
	}) 
}

func TestMustHaveTag(t *testing.T) {

	t.Run("Exists", func(t *testing.T){
		type sFive struct {
			FieldOne string `bson:"tag1"`
		}
		assert.NotEmpty(t, MustHaveTag(sFive{}, "FieldOne"))
	})
	t.Run("IsEmpty", func(t *testing.T) {
		type sSix struct{
			FieldOne string `"foo"`
		}
		assert.Panics(t, func() {MustHaveTag(sSix{}, "FieldOne")})
	})
	t.Run("Errors", func(t *testing.T){
		type sSeven struct{
			FieldOne string `bson:"tag1"`
		}
		assert.Panics(t, func() {MustHaveTag(sSeven{}, "FieldTwo")})
	})
	
}
