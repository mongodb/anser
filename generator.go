package anser

import (
	"github.com/mongodb/amboy"
	"github.com/mongodb/grip"
)

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
	// Returns the list of dependencies (e.g. edges) that all jobs
	// produced by this generator should have.
	Dependencies() []string

	// Jobs produces job objects for the results of the
	// generator.
	Jobs() <-chan amboy.Job

	// MigrationGenerators are themselves amboy.Jobs.
	amboy.Job
}

// AddMigrationJobs takes an amboy.Queue, processes the results, and
// adds any jobs produced by the generator, with the appropriate
// dependencies
func AddMigrationJobs(q amboy.Queue) (int, error) {
	catcher := grip.NewCatcher()
	count := 0
	for job := range q.Results() {
		generator, ok := job.(MigrationGenerator)
		if !ok {
			continue
		}
		grip.Infof("adding operations for %s", generator.ID())

		for j := range generator.Jobs() {
			dep := j.Dependency()
			for _, edge := range generator.Dependencies() {
				catcher.Add(dep.Add(edge))
			}
			// this is almost certainly unnecessary in
			// most cases.
			j.SetDependency(dep)

			catcher.Add(q.Put(j))
		}

		count++
	}

	grip.Info("added %d migration operations", count)
	return count, catcher.Resolve()
}
