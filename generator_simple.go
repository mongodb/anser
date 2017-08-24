package anser

import (
	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/registry"
	"github.com/mongodb/grip"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	registry.AddJobType("simple-migration-generator",
		func() amboy.Job { return makeSimpleGenerator() })
}

func NewSimpleMigrationGenerator(e Environment, ns Namespace,
	query, update map[string]interface{}) amboy.Job {

	j := makeSimpleGenerator()
	j.MigrationHelper = MigrationHelper(e)
	j.NS = ns
	j.Query = query
	j.Update = update

	return j
}

func makeSimpleGenerator() *simpleMigrationGenerator {
	return &simpleMigrationGenerator{
		MigrationHelper: &migrationBase{},
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    "simple-migration-generator",
				Version: 0,
				Format:  amboy.BSON,
			},
		},
	}
}

type simpleMigrationGenerator struct {
	NS              Namespace              `bson:"ns" json:"ns" yaml:"ns"`
	Query           map[string]interface{} `bson:"source_query" json:"source_query" yaml:"source_query"`
	Update          map[string]interface{} `bson:"update" json:"update" yaml:"update"`
	job.Base        `bson:"job_base" json:"job_base" yaml:"job_base"`
	MigrationHelper `bson:"-" json:"-" yaml:"-"`
}

func (j *simpleMigrationGenerator) Run() {
	defer j.MarkComplete()

	env := j.Env()
	session, err := j.GetSession()
	if err != nil {
		j.AddError(err)
		return
	}
	defer session.Close()

	queue, err := j.GetQueue()
	if err != nil {
		j.AddError(err)
		return
	}

	coll := session.DB(NS.DB).C(NS.Collection)
	iter := coll.Find(j.Query).Select(bson.M{"_id": 1}).Iter()

	doc := struct {
		ID interface{} `bson:"_id"`
	}{}

	catcher := grip.NewCatcher()

	for iter.Next(&doc) {
		m := NewSimpleMigration(env, SimpleMigration{
			ID:        doc.ID,
			Update:    Update,
			Migration: j.ID(),
			Namespace: j.NS,
		})

		// TODO setup dependencies here
		catcher.Add(queue.Put(m))
	}

	if catcher.HasErrors() {
		j.AddError(catcher.Resolve())
		return
	}

	if err := iter.Close(); err != nil {
		j.AddError(err)
		return
	}
}
