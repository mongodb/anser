package anser

import (
	"errors"
	"sync"
)

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
