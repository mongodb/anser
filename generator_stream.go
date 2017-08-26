package anser

import (
	"fmt"
	"sync"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/registry"
	"github.com/mongodb/grip"
	"github.com/tychoish/anser/model"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	registry.AddJobType("stream-migration-generator",
		func() amboy.Job { return makeStreamGenerator() })
}

func NewStreamMigrationGenerator(e Environment, opts GeneratorOptions, opName string) Generator {
	j := makeStreamGenerator()
	j.SetID(opts.JobID)
	j.SetDependency(opts.dependency())
	j.MigrationHelper = NewMigrationHelper(e)
	j.NS = opts.NS
	j.Query = opts.Query
	j.ProcessorName = opName

	return j
}

func makeStreamGenerator() *streamMigrationGenerator {
	return &streamMigrationGenerator{
		MigrationHelper: &migrationBase{},
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    "stream-migration-generator",
				Version: 0,
				Format:  amboy.BSON,
			},
		},
	}
}

type streamMigrationGenerator struct {
	NS              model.Namespace        `bson:"ns" json:"ns" yaml:"ns"`
	Query           map[string]interface{} `bson:"source_query" json:"source_query" yaml:"source_query"`
	ProcessorName   string                 `bson:"processor_name" json:"processor_name" yaml:"processor_name"`
	Migrations      []*streamMigrationJob  `bson:"migrations" json:"migrations" yaml:"migrations"`
	job.Base        `bson:"job_base" json:"job_base" yaml:"job_base"`
	MigrationHelper `bson:"-" json:"-" yaml:"-"`
	mu              sync.Mutex
}

func (j *streamMigrationGenerator) Run() {
	defer j.MarkComplete()

	env := j.Env()

	network, err := env.GetDependencyNetwork()
	if err != nil {
		j.AddError(err)
		return
	}

	session, err := env.GetSession()
	if err != nil {
		j.AddError(err)
		return
	}
	defer session.Close()

	coll := session.DB(j.NS.DB).C(j.NS.Collection)
	iter := coll.Find(j.Query).Select(bson.M{"_id": 1}).Iter()

	doc := struct {
		ID interface{} `bson:"_id"`
	}{}

	ids := []string{}
	j.mu.Lock()
	defer j.mu.Unlock()
	for iter.Next(&doc) {
		m := NewStreamMigration(env, model.Stream{
			// ID:            doc.ID,
			ProcessorName: j.ProcessorName,
			Migration:     j.ID(),
			Namespace:     j.NS,
		}).(*streamMigrationJob)

		dep, err := env.NewDependencyManager(j.ID(), j.Query, j.NS)
		if err != nil {
			j.AddError(err)
			continue
		}
		m.SetDependency(dep)
		m.SetID(fmt.Sprintf("%s.%v.%d", j.ID(), doc.ID, len(ids)))
		ids = append(ids, m.ID())
		j.Migrations = append(j.Migrations, m)
	}

	network.AddGroup(j.ID(), ids)

	if err := iter.Close(); err != nil {
		j.AddError(err)
		return
	}
}

func (j *streamMigrationGenerator) Jobs() <-chan amboy.Job {
	env := j.Env()

	j.mu.Lock()
	defer j.mu.Unlock()

	jobs := make(chan amboy.Job, len(j.Migrations))
	for _, job := range j.Migrations {
		jobs <- job
	}
	close(jobs)

	out, err := generator(env, j.ID(), jobs)
	grip.CatchError(err)
	grip.Infof("produced %d tasks for migration %s", len(j.Migrations), j.ID())
	j.Migrations = []*streamMigrationJob{}
	return out
}
