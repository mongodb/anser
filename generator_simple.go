package anser

import (
	"sync"

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
	Migrations      []*simpleMigrationJob  `bson:"migrations" json:"migrations" yaml:"migrations"`
	job.Base        `bson:"job_base" json:"job_base" yaml:"job_base"`
	MigrationHelper `bson:"-" json:"-" yaml:"-"`
	mu              sync.Mutex
}

func (j *simpleMigrationGenerator) Run() {
	defer j.MarkComplete()

	env := j.Env()

	network, err := env.GetDependencyNetwork()
	if err != nil {
		j.AddError(err)
		return
	}

	session, err := j.GetSession()
	if err != nil {
		j.AddError(err)
		return
	}
	defer session.Close()

	coll := session.DB(NS.DB).C(NS.Collection)
	iter := coll.Find(j.Query).Select(bson.M{"_id": 1}).Iter()

	doc := struct {
		ID interface{} `bson:"_id"`
	}{}

	ids := []string{}
	for iter.Next(&doc) {
		m := NewSimpleMigration(env, SimpleMigration{
			ID:        doc.ID,
			Update:    Update,
			Migration: j.ID(),
			Namespace: j.NS,
		}).(*simpleMigrationJob)

		m.SetID("<>")
		ids = append(ids, m.ID())
		j.Migrations = append(j.Migrations, m)
	}

	network.AddGroup(j.ID(), ids)

	if err := iter.Close(); err != nil {
		j.AddError(err)
		return
	}
}

func (j *simpleMigrationGenerator) Jobs() <-chan amboy.Job {
	out := make(chan amboy.Job, len(j.Migrations))
	for idx, migration := range j.Migrations {

	}
	grip.Info("produced %s generated ")
}
