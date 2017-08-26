/*
Migration Execution Environment

Anser provides the Environment interface, with a global instance
accessible via the exported GetEnvironment() function to provide
access to runtime configuration state: database connections;
amboy.Queue objects, and registries for task implementations.

The Environment is an interface: you can build a mock, or use one
provided for testing purposes by anser (coming soon).

*/
package anser

import (
	"sync"

	"github.com/mongodb/amboy"
	"github.com/pkg/errors"
	mgo "gopkg.in/mgo.v2"
)

const (
	defaultMetadataCollection = "migrations.metadata"
	defaultAnserDB            = "anser"
)

var globalEnv *envState

func init() {
	globalEnv = &envState{
		migrations: make(map[string]ManualMigrationOperation),
		processor:  make(map[string]DocumentProcessor),
	}
}

// Environment exposes the execution environment for the migration
// utility, and is the method by which, potentially serialized job
// definitions are able to gain access to the database and through
// which generator jobs are able to gain access to the queue.
//
// Implementations should be thread-safe, and are not required to be
// reconfigurable after their initial configuration.
type Environment interface {
	Setup(amboy.Queue, string) error
	GetSession() (*mgo.Session, error)
	GetQueue() (amboy.Queue, error)
	GetDependencyNetwork() (DependencyNetworker, error)
	MetadataNamespace() Namespace
	RegisterManualMigrationOperation(string, ManualMigrationOperation) error
	GetManualMigrationOperation(string) (ManualMigrationOperation, bool)
	RegisterDocumentProcessor(string, DocumentProcessor) error
	GetDocumentProcessor(string) (DocumentProcessor, bool)
}

// GetEnvironment returns the global environment object. Because this
// produces a pointer to the global object, make sure that you have a
// way to replace it with a mock as needed for testing.
func GetEnvironment() Environment { return globalEnv }

type envState struct {
	queue      amboy.Queue
	session    *mgo.Session
	metadataNS Namespace
	deps       DependencyNetworker
	migrations map[string]ManualMigrationOperation
	processor  map[string]DocumentProcessor
	isSetup    bool
	mu         sync.RWMutex
}

func (e *envState) Setup(q amboy.Queue, mongodbURI string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isSetup {
		return errors.New("reconfiguring a queue is not supported")
	}

	session, err := mgo.Dial(mongodbURI)
	if err != nil {
		return errors.Wrap(err, "problem establishing connection")
	}

	if !q.Started() {
		return errors.New("configuring anser environment with a non-running queue")
	}

	if session.DB("").Name != "test" {
		e.metadataNS.DB = defaultAnserDB
	}

	e.queue = q
	e.session = session
	e.metadataNS.Collection = defaultMetadataCollection
	e.isSetup = true
	e.deps = NewDependencyNetwork()

	return nil
}

func (e *envState) GetSession() (*mgo.Session, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.session == nil {
		return nil, errors.New("no session defined")
	}

	return e.session.Clone(), nil
}

func (e *envState) GetQueue() (amboy.Queue, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.queue == nil {
		return nil, errors.New("no queue defined")
	}

	return e.queue, nil
}

func (e *envState) GetDependencyNetwork() (DependencyNetworker, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.deps == nil {
		return nil, errors.New("no dependency networker specified")
	}

	return e.deps, nil
}

func (e *envState) RegisterManualMigrationOperation(name string, op ManualMigrationOperation) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.migrations[name]; ok {
		return errors.Errorf("migration operation %s already exists", name)
	}

	e.migrations[name] = op
	return nil
}

func (e *envState) GetManualMigrationOperation(name string) (ManualMigrationOperation, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	op, ok := e.migrations[name]
	return op, ok
}

func (e *envState) RegisterDocumentProcessor(name string, docp DocumentProcessor) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.processor[name]; ok {
		return errors.Errorf("document processor named %s already registered", name)
	}

	e.processor[name] = docp
	return nil
}

func (e *envState) GetDocumentProcessor(name string) (DocumentProcessor, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	docp, ok := e.processor[name]
	return docp, ok
}

func (e *envState) MetadataNamespace() Namespace {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.metadataNS
}
