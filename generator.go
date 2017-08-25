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
	// Jobs produces job objects for the results of the
	// generator.
	Jobs() <-chan amboy.Job

	// MigrationGenerators are themselves amboy.Jobs.
	amboy.Job
}

// AddMigrationJobs takes an amboy.Queue, processes the results, and
// adds any jobs produced by the generator to the queue.
func AddMigrationJobs(q amboy.Queue, dryRun bool) (int, error) {
	catcher := grip.NewCatcher()
	count := 0
	for job := range q.Results() {
		generator, ok := job.(MigrationGenerator)
		if !ok {
			continue
		}
		grip.Infof("adding operations for %s", generator.ID())

		for j := range generator.Jobs() {
			if dryRun {
				catcher.Add(q.Put(j))
			}
		}

		count++
	}

	grip.Info("added %d migration operations", count)
	return count, catcher.Resolve()
}

// generator provides the high level implementation of the Jobs()
// method that's a part of the MigrationGenerator interface. This
// takes a list of jobs (using a variadic function to do the type
// conversion,) and returns them in a (buffered) channel. with the
// jobs, having had their dependencies set.
func generator(env Environment, groupID string, migrations ...amboy.Job) (<-chan amboy.Job, error) {
	network, err := env.GetDependencyNetwork()
	if err != nil {
		grip.Warning(err)
		return
	}

	out := make(chan amboy.Job, len(migrations))

	for _, migration := range migrations {
		dep := migration.Dependency()

		for _, group := range network.Resolve(groupID) {
			for _, edge := range network.GetGroup(group) {
				grip.CatchNotice(dep.AddEdge(edge))
			}
		}

		migration.SetDependency(dep)

		out <- migration
	}

	return out
}
