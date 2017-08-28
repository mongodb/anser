package anser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2/bson"
)

type DependencyManagerSuite struct {
	dep *migrationDependency
	suite.Suite
}

func TestDependencyManagerSuite(t *testing.T) {
	suite.Run(t, new(DependencyManagerSuite))
}

func (s *DependencyManagerSuite) SetupTest() {
	s.dep = makeMigrationDependencyManager()
}

func (s *DependencyManagerSuite) TestDefaultTypeInfo() {
	s.Zero(s.dep.MigrationID)
	s.Equal(s.dep.Type().Name, "anser-migration")
}

func TestDependencyStateQuery(t *testing.T) {
	assert := assert.New(t)
	keys := []string{"foo", "bar"}
	query := getDependencyStateQuery(keys)

	assert.Len(query, 1)
	idClause, ok := query["_id"]
	assert.True(ok)
	assert.Len(idClause, 1)
	inClause := idClause.(bson.M)["$in"].([]string)
	assert.Len(inClause, 2)
}

// func (s *DependencyManagerSuite) Test() {}
// func (s *DependencyManagerSuite) Test() {}
