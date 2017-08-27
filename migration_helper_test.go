package anser

import (
	"testing"

	"github.com/mongodb/amboy/queue"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

type MigrationHelperSuite struct {
	env    *envState
	mh     *migrationBase
	cancel context.CancelFunc
	suite.Suite
}

func TestMigrationHelperSuite(t *testing.T) {
	suite.Run(t, new(MigrationHelperSuite))
}

func (s *MigrationHelperSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	q := queue.NewLocalUnordered(4)
	s.env = &envState{
		migrations: make(map[string]ManualMigrationOperation),
		processor:  make(map[string]DocumentProcessor),
	}

	s.NoError(q.Start(ctx))
	s.NoError(s.env.Setup(q, "mongodb://localhost:27017/"))
}

func (s *MigrationHelperSuite) TearDownSuite() {
	s.cancel()
}

func (s *MigrationHelperSuite) SetupTest() {
	s.mh = NewMigrationHelper(s.env).(*migrationBase)
}

func (s *MigrationHelperSuite) TestEnvironmentConfiguredByDefault() {
	s.Equal(s.env, s.mh.Env())
	s.Equal(s.env, s.mh.env)
}

// TODO need a mock env setup to test MigrationHelper functionality
