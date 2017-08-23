package anser

import (
	"sync"

	"github.com/mongodb/amboy/job"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

type MigrationMetadata struct {
	ID        string `bson:"_id" json:"id" yaml:"id"`
	Migration string `bson:"migration" json:"migration" yaml:"migration"`
	HasErrors bool   `bson:"has_errors" json:"has_errors" yaml:"has_errors"`
	Completed bool   `bson:"completed" json:"completed" yaml:"completed"`
}

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

	SaveMigrationEvent(*MigrationMetadata) error
	FinishMigration(string, *job.Base)
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
