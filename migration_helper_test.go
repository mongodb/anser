package anser

import (
	"testing"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tychoish/anser/mock"
	"golang.org/x/net/context"
)

type MigrationHelperSuite struct {
	env    *mock.Environment
	mh     *migrationBase
	queue  amboy.Queue
	cancel context.CancelFunc
	suite.Suite
}

func TestMigrationHelperSuite(t *testing.T) {
	suite.Run(t, new(MigrationHelperSuite))
}

func (s *MigrationHelperSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.queue = queue.NewLocalUnordered(4)

	s.env = mock.NewEnvironment()
	s.NoError(s.queue.Start(ctx))
	s.NoError(s.env.Setup(s.queue, "mongodb://localhost:27017/"))
}

func (s *MigrationHelperSuite) TearDownSuite() {
	s.cancel()
}

func (s *MigrationHelperSuite) SetupTest() {
	s.env = mock.NewEnvironment()
	s.env.Queue = s.queue
	s.mh = NewMigrationHelper(s.env).(*migrationBase)
}

// TODO need a mock env setup to test MigrationHelper functionality

func TestDefaultEnvironmentAndMigrationHelperState(t *testing.T) {
	assert := assert.New(t)
	env := &envState{}
	mh := NewMigrationHelper(env).(*migrationBase)
	assert.Equal(env, mh.Env())
	assert.Equal(env, mh.env)

	assert.Equal(globalEnv, GetEnvironment())
}
