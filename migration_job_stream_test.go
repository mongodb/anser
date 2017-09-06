package anser

import (
	"testing"

	"github.com/mongodb/amboy/registry"
	"github.com/stretchr/testify/assert"
	"github.com/tychoish/anser/mock"
)

func TestStreamMigrationJob(t *testing.T) {
	assert := assert.New(t)
	const jobTypeName = "stream-migration"
	env := mock.NewEnvironment()
	mh := &MigrationHelperMock{Environment: env}

	// first test the factory
	factory, err := registry.GetJobFactory(jobTypeName)
	assert.NoError(err)
	job, ok := factory().(*streamMigrationJob)
	assert.True(ok)
	assert.Equal(job.Type().Name, jobTypeName)

	// now we run the jobs a couple of times to verify expected behavior and outcomes

	// first we start off with a "processor not defined error"
	job.MigrationHelper = mh
	job.Definition.ProcessorName = "processor-name"
	job.Run()
	assert.True(job.HasErrors())
	err = job.Error()
	assert.Error(err)
	assert.Contains(err.Error(), job.Definition.ProcessorName)

	// set up the processor
	processor := &mock.Processor{}
	env.ProcessorRegistry["processor-name"] = processor

	// reset for next case.
	// now we find a poorly configured/implemented processor
	job = factory().(*streamMigrationJob)
	job.Definition.ProcessorName = "processor-name"
	processor.Iter = &mock.Iterator{}
	job.Run()
	assert.True(job.HasErrors())

	// reset for next case.
	job = factory().(*streamMigrationJob)
	job.MigrationHelper = mh
	job.Definition.ProcessorName = "processor-name"
	job.Run()
	assert.False(job.HasErrors())

}
