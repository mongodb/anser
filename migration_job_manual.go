package anser

import (
	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/registry"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	registry.AddJobType("manual-migration", func() amboy.Job { return makeManualMigration() })
}

func NewManualMigration(e Environment, m ManualMigration) amboy.Job {
	j := makeManualMigration()
	j.MigrationHelper = NewMigrationHelper(e)
	j.Definition = m
	return j
}

func makeManualMigration() *manualMigrationJob {
	return &manualMigrationJob{
		MigrationHelper: &migrationBase{},
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    "manual-migration",
				Version: 0,
				Format:  amboy.BSON,
			},
		},
	}
}

type manualMigrationJob struct {
	Definition      ManualMigration `bson:"migration" json:"migration" yaml:"migration"`
	job.Base        `bson:"job_base" json:"job_base" yaml:"job_base"`
	MigrationHelper `bson:"-" json:"-" yaml:"-"`
}

func (j *manualMigrationJob) Run() {
	defer j.MarkComplete()

	env := j.Env()

	operation, ok := env.GetManualMigrationOperation(j.Definition.OperationName)
	if !ok {
		j.AddError(errors.Errorf("could not find migration operation named %s", j.Definition.OperationName))
		return
	}

	session, err := env.GetSession()
	if err != nil {
		j.AddError(errors.Wrap(err, "problem getting database session"))
		return
	}
	defer session.Close()

	var doc bson.Raw
	coll := session.DB(j.Definition.Namespace.DB).C(j.Definition.Namespace.Collection)
	err = coll.FindId(j.Definition.ID).One(&doc)
	if err != nil {
		j.AddError(err)
		return
	}

	j.AddError(operation(session, doc))
}
