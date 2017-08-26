package anser

import (
	"sync"

	"github.com/mongodb/amboy/job"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"github.com/tychoish/anser/model"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MigrationMetadata struct {
	ID        string `bson:"_id" json:"id" yaml:"id"`
	Migration string `bson:"migration" json:"migration" yaml:"migration"`
	HasErrors bool   `bson:"has_errors" json:"has_errors" yaml:"has_errors"`
	Completed bool   `bson:"completed" json:"completed" yaml:"completed"`
}

// Satisfies reports if a migration has completed without errors.
func (m *MigrationMetadata) Satisfied() bool { return m.Completed && !m.HasErrors }

// MigrationHelper is an interface embedded in all jobs as an
// "extended base" for migrations ontop for common functionality of
// the existing amboy.Base type which implements most job
// functionality.
//
// MigrationHelper implementations should not require construction:
// getter methods should initialize nil values at runtime.
type MigrationHelper interface {
	Env() Environment
	SetEnv(Environment) error

	// Migrations need to record their state to help resolve
	// dependencies to the database.
	FinishMigration(string, *job.Base)
	SaveMigrationEvent(*MigrationMetadata) error

	// The migration helper provides a model/interface for
	// interacting with the database to check the state of a
	// migration operation, helpful in dependency approval.
	PendingMigrationOperations(model.Namespace, map[string]interface{}) int
	GetMigrationEvents(map[string]interface{}) (*mgo.Session, DocumentIterator, error)
}

// NewMigrationHelper constructs a new migration helper instance. Use
// this to inject environment instances into tasks.
func NewMigrationHelper(e Environment) MigrationHelper { return &migrationBase{env: e} }

type migrationBase struct {
	env Environment
	sync.Mutex
}

func (e *migrationBase) Env() Environment {
	e.Lock()
	defer e.Unlock()

	if e.env == nil {
		e.SetEnv(GetEnvironment())
	}

	return e.env
}

func (e *migrationBase) SetEnv(en Environment) error {
	if en == nil {
		return errors.New("cannot set environment to nil")
	}

	e.Lock()
	defer e.Unlock()
	e.env = en

	return nil
}

func (e *migrationBase) SaveMigrationEvent(m *MigrationMetadata) error {
	env := e.Env()

	session, err := env.GetSession()
	if err != nil {
		return errors.WithStack(err)
	}
	defer session.Close()

	ns := env.MetadataNamespace()
	coll := session.DB(ns.DB).C(ns.Collection)
	_, err = coll.UpsertId(m.ID, m)
	return errors.Wrap(err, "problem inserting migration metadata")
}

func (e *migrationBase) FinishMigration(name string, j *job.Base) {
	j.MarkComplete()
	meta := MigrationMetadata{
		ID:        j.ID(),
		Migration: name,
		HasErrors: j.HasErrors(),
		Completed: true,
	}
	err := e.SaveMigrationEvent(&meta)
	if err != nil {
		j.AddError(err)
		grip.Warningf("encountered problem [%s] saving migration metadata", err.Error())
	}
}

func (e *migrationBase) PendingMigrationOperations(ns model.Namespace, q map[string]interface{}) int {
	env := e.Env()

	session, err := env.GetSession()
	if err != nil {
		grip.Error(errors.WithStack(err))
		return -1
	}
	defer session.Close()

	coll := session.DB(ns.DB).C(ns.Collection)

	num, err := coll.Find(bson.M(q)).Count()
	if err != nil {
		grip.Warning(errors.WithStack(err))
		return -1
	}

	return num
}

func (e *migrationBase) GetMigrationEvents(q map[string]interface{}) (*mgo.Session, DocumentIterator, error) {
	env := e.Env()
	ns := env.MetadataNamespace()

	session, err := env.GetSession()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	iter := session.DB(ns.DB).C(ns.Collection).Find(bson.M(q)).Iter()
	return session, iter, nil
}
