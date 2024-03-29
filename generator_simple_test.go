package anser

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mongodb/amboy/registry"
	"github.com/mongodb/anser/mock"
	"github.com/mongodb/anser/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleMigrationGenerator(t *testing.T) {
	ctx := context.Background()
	env := mock.NewEnvironment()
	mh := &MigrationHelperMock{Environment: env}
	const jobTypeName = "simple-migration-generator"

	factory, err := registry.GetJobFactory(jobTypeName)
	require.NoError(t, err)
	job, ok := factory().(*simpleMigrationGenerator)
	require.True(t, ok)
	require.Equal(t, job.Type().Name, jobTypeName)

	ns := model.Namespace{DB: "foo", Collection: "bar"}
	opts := model.GeneratorOptions{}

	t.Run("Interface", func(t *testing.T) {
		assert.Implements(t, (*Generator)(nil), &simpleMigrationGenerator{})
	})
	t.Run("Constructor", func(t *testing.T) {
		// check that the public method produces a reasonable object
		// of the correct type, without shared state
		generator := NewSimpleMigrationGenerator(env, opts, nil).(*simpleMigrationGenerator)
		require.NotNil(t, generator)
		assert.Equal(t, generator.Type().Name, jobTypeName)
		assert.NotEqual(t, generator, job)
	})
	t.Run("DependencyCheck", func(t *testing.T) {
		// check that the run method returns an error if it can't get a dependency error
		env.NetworkError = errors.New("injected network error")
		job.MigrationHelper = mh
		job.Run(ctx)
		assert.True(t, job.Status().Completed)
		require.True(t, job.HasErrors())
		err = job.Error()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "injected network error")
		env.NetworkError = nil
	})
	t.Run("Client", func(t *testing.T) {
		env.Client = mock.NewClient()
		defer func() {
			env.Client = nil
		}()
		t.Run("ClientError", func(t *testing.T) {
			// make sure that client acquisition errors propagate
			job = factory().(*simpleMigrationGenerator)
			env.ClientError = errors.New("injected client error")
			job.MigrationHelper = mh
			job.Run(ctx)
			assert.True(t, job.Status().Completed)
			if assert.True(t, job.HasErrors()) {
				err = job.Error()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "injected client error")
			}
			env.ClientError = nil

		})
		t.Run("FindError", func(t *testing.T) {
			// make sure that client acquisition errors propagate
			job = factory().(*simpleMigrationGenerator)
			job.NS.DB = "foo"
			job.NS.Collection = "bar"
			assert.NotNil(t, env.Client.Database("foo").Collection("bar"))
			env.Client.Databases["foo"].Collections["bar"].FindError = errors.New("injected query error")
			job.MigrationHelper = mh
			job.Run(ctx)
			assert.True(t, job.Status().Completed)
			if assert.True(t, job.HasErrors()) {
				err = job.Error()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "injected query error")
			}
			env.ClientError = nil
		})
		t.Run("Generation", func(t *testing.T) {
			defer func() { env.Network = mock.NewDependencyNetwork() }()
			env.Network = mock.NewDependencyNetwork()
			// make sure that we generate the jobs we would expect to:
			job = factory().(*simpleMigrationGenerator)
			job.NS = ns
			job.MigrationHelper = mh
			job.Limit = 3
			job.SetID("simple")

			cursor := &mock.Cursor{
				Results: []interface{}{
					&doc{"one"}, &doc{"two"}, &doc{"three"}, &doc{"four"},
				},
				ShouldIter:   true,
				MaxNextCalls: 4,
			}

			ids := job.generateJobs(ctx, env, cursor)
			for idx, id := range ids {
				require.True(t, strings.HasPrefix(id, "simple."))
				require.True(t, strings.HasSuffix(id, fmt.Sprintf(".%d", idx)))
				switch idx {
				case 0:
					assert.Contains(t, id, ".one.")
				case 1:
					assert.Contains(t, id, ".two.")
				case 2:
					assert.Contains(t, id, ".three.")
				case 3:
					assert.Contains(t, id, ".four.")
				}
			}

			assert.Len(t, ids, 3)
			assert.Len(t, job.Migrations, 3)

			network, err := env.GetDependencyNetwork()
			require.NoError(t, err)
			network.AddGroup(job.ID(), ids)
			networkMap := network.Network()
			assert.Len(t, networkMap[job.ID()], 3)
		})

	})
}
