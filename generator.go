package anser

import "github.com/mongodb/amboy"

// MigrationGenerator is a amboy.Job super set used to store
// implementations that generate migration jobs. Internally they
// create large jobs.
//
// The current limitation is that the generated jobs are stored within
// the generator job, which means they must either all fit in memory
// *or* be serializeable independently (e.g. fit in the 16mb document
// limit if using a MongoDB backed query.)
//
// Indeed this interface may be useful at some point for any kind of
// job generating operation.
type MigrationGenerator interface {
	Dependencies() []string
	Jobs() <-chan amboy.Job

	amboy.Job
}
