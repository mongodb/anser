package anser

import (
	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/registry"
	"github.com/pkg/errors"
)

func init() {
	registry.AddJobType("aggregate-migration", func() amboy.Job { return makeAggregateProducer() })
}

func NewAggregateMigration(e Environment, m AggregateProducer) amboy.Job {
	j := makeAggregateProducer()
	j.MigrationHelper = NewMigrationHelper(e)
	j.Definition = m
	return j
}

func makeAggregateProducer() *aggregateMigrationJob {
	return &aggregateMigrationJob{
		MigrationHelper: &migrationBase{},
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    "aggregate-migration",
				Version: 0,
				Format:  amboy.BSON,
			},
		},
	}
}

type aggregateMigrationJob struct {
	Definition      AggregateProducer `bson:"migration" json:"migration" yaml:"migration"`
	job.Base        `bson:"job_base" json:"job_base" yaml:"job_base"`
	MigrationHelper `bson:"-" json:"-" yaml:"-"`
}

func (j *aggregateMigrationJob) Run() {
	defer j.MarkComplete()

	env := j.Env()

	producer, ok := env.GetDocumentProcessor(j.Definition.ProcessorName)
	if !ok {
		j.AddError(errors.Errorf("producer named %s is not defined",
			j.Definition.ProcessorName))
		return
	}

	iter := producer.Load(j.Definition.Namespace, j.Definition.Query)
	if iter == nil {
		j.AddError(errors.Errorf("document processor for %s could not return iterator",
			j.Definition.Migration))
		return
	}

	j.AddError(producer.Migrate(iter))
}
