package anser

import (
	"fmt"
	"testing"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/queue"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tychoish/anser/db"
	"github.com/tychoish/anser/mock"
	"github.com/tychoish/anser/model"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
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
	s.NoError(s.queue.Start(ctx))
}

func (s *MigrationHelperSuite) TearDownSuite() {
	s.cancel()
}

func (s *MigrationHelperSuite) SetupTest() {
	s.env = mock.NewEnvironment()
	s.env.MetaNS = model.Namespace{"anserDB", "anserMeta"}
	s.env.Queue = s.queue
	s.mh = NewMigrationHelper(s.env).(*migrationBase)
	s.NoError(s.env.Setup(s.queue, "mongodb://localhost:27017/"))
}

func (s *MigrationHelperSuite) TestEnvironmentIsConsistent() {
	s.Equal(s.mh.Env(), s.env)
	s.NotEqual(s.mh.Env(), globalEnv)
}

func printDebug(v interface{}) { fmt.Printf("%T: %+v\n", v, v) }

func (s *MigrationHelperSuite) TestSaveMigrationEvent() {
	s.env.SessionError = errors.New("session error")
	err := errors.Cause(s.mh.SaveMigrationEvent(nil))
	s.Error(err)
	s.Equal(err, s.env.SessionError)
	s.env.SessionError = nil

	err = s.mh.SaveMigrationEvent(&model.MigrationMetadata{})
	s.NoError(err)

	db := s.env.Session.DBs["anserDB"]
	s.NotNil(db)
	coll, ok := db.Collections["anserMeta"]
	s.True(ok)
	s.NotNil(coll)
	s.Len(coll.InsertedDocs, 1)
	coll.FailWrites = true
	err = s.mh.SaveMigrationEvent(&model.MigrationMetadata{})
	s.Error(err)
	s.Equal(errors.Cause(err).Error(), "writes fail")
	s.Len(coll.InsertedDocs, 1)
}

func (s *MigrationHelperSuite) TestFinishMigrationIsTracked() {
	base := &job.Base{}

	status := base.Status()
	s.False(status.Completed)

	s.mh.FinishMigration("foo", base)

	status = base.Status()
	s.True(status.Completed)

	db := s.env.Session.DBs["anserDB"]
	s.NotNil(db)
	coll, ok := db.Collections["anserMeta"]
	s.True(ok)
	s.NotNil(coll)
	s.Len(coll.InsertedDocs, 1)
	doc, ok := coll.InsertedDocs[0].(*model.MigrationMetadata)
	s.True(ok)
	s.Equal(doc.Migration, "foo")
}

func (s *MigrationHelperSuite) TestGetMigrationEvents() {
	s.env.SessionError = errors.New("session error")
	query := map[string]interface{}{"foo": 1}

	iter, err := s.mh.GetMigrationEvents(query)
	s.Nil(iter)
	s.Error(err)
	s.Equal(errors.Cause(err), s.env.SessionError)
	s.env.SessionError = nil

	iter, err = s.mh.GetMigrationEvents(query)
	mi := iter.(db.CombinedCloser).Iterator.(*mock.Iterator)
	s.NotNil(iter)
	s.NoError(err)
	s.Equal(mi.Query.Query, bson.M(query))
	coll, ok := s.env.Session.DBs["anserDB"].Collections["anserMeta"]
	s.True(ok)
	s.NotNil(coll)
	s.Len(coll.Queries, 1)
}

func TestDefaultEnvironmentAndMigrationHelperState(t *testing.T) {
	assert := assert.New(t)
	env := &envState{}
	mh := NewMigrationHelper(env)
	assert.Equal(env, mh.Env())
	assert.Equal(env, mh.(*migrationBase).env)

	assert.Equal(globalEnv, GetEnvironment())
}
