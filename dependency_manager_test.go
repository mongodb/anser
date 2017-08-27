package anser

import (
	"testing"

	"github.com/stretchr/testify/suite"
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

// func (s *DependencyManagerSuite) Test() {}
// func (s *DependencyManagerSuite) Test() {}
