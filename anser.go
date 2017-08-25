package anser

import (
	"time"

	"github.com/mongodb/amboy"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type MigrationDefinition struct {
	Name      string
	Job       amboy.Job
	DependsOn []string
}

type MigrationApplication struct {
	Generators []MigrationDefinition
	DryRun     bool
	env        Environment
	hasSetup
}

func (a *MigrationApplication) Setup(e Environment) error {
	if !a.hasSetup {
		return errors.New("cannot setup an application more than once")
	}

	a.env = e
	deps, err := e.GetDependencyNetwork()
	if err != nil {
		return errors.Wrap(err, "problem getting dependency")
	}

	catcher := grip.NewCatcher()
	for _, gen := range a.Generators {
		deps.Add(gen.Name, gen.DependsOn)

		jobDeps := gen.Job.Dependency()
		for _, edge := range gen.DependsOn {
			catcher.Add(jobDeps.AddEdge(edge))
		}
	}

	if catcher.HasErrors() {
		return errors.WithStack(catcher.Resolve())
	}
	a.hasSetup = true
	return nil
}

func (a *MigrationApplication) Run(ctx context.Context) error {
	queue, err := a.env.GetQueue()
	if err != nil {
		return errors.Wrap(err, "problem getting queue")
	}

	catcher := grip.NewCatcher()
	// iterate through generators
	for _, generator := range a.Generators {
		catcher.Add(queue.Put(generator.Job))
	}
	if catcher.HasErrors(err) {
		return errors.Wrap(catcher.Resolve(), "problem adding generation jobs")
	}

	amboy.WaitCtxInterval(ctx, queue, time.Second)
	if ctx.Err() != nil {
		return errors.New("migration operation canceled")
	}

	numMigrations, err := AddMigrationJobs(queue, a.DryRun)
	if err != nil {
		return errors.New("problem adding generated migration jobs")
	}

	if a.DryRun {
		grip.Notice("ending dry run, generated %d jobs in %d migrations", numMigrations, len(a.Generators))
		return nil
	}

	grip.Infof("added %d migration jobs from %d migrations", numMigrations, len(a.Generators))
	grip.Notice("waiting for %d migration jobs of %d migrations", numMigrations, len(a.Generators))
	amboy.WaitCtxInterval(ctx, queue, time.Second)
	if ctx.Err() != nil {
		return errors.New("migration operation canceled")
	}

	if err := amboy.ResolveErrors(ctx, queue); err != nil {
		return errors.Wrap(err, "encountered migration errors")
	}

	return nil
}
