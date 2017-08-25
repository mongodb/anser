package anser

import (
	"fmt"
	"sync"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/registry"
	"github.com/mongodb/grip"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	registry.AddJobType("manual-migration-generator",
		func() amboy.Job { return makeManualGenerator() })
}

func NewSimpleMigrationGenerator(e Environment, ns Namespace,
	query map[string]interface{}, opName string) amboy.Job {

	j := makeManualGenerator()
	j.MigrationHelper = MigrationHelper(e)
	j.NS = ns
	j.Query = query
	j.OperationName = opName

	return j
}

func makeManualGenerator() *manualMigrationGenerator {
	return &manualMigrationGenerator{
		MigrationHelper: &migrationBase{},
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    "manual-migration-generator",
				Version: 0,
				Format:  amboy.BSON,
			},
		},
	}
}

type manualMigrationGenerator struct {
	NS              Namespace              `bson:"ns" json:"ns" yaml:"ns"`
	Query           map[string]interface{} `bson:"source_query" json:"source_query" yaml:"source_query"`
	OperationName   string                 `bson:"op_name" json:"op_name" yaml:"op_name"`
	Migrations      []*manualMigrationJob  `bson:"migrations" json:"migrations" yaml:"migrations"`
	job.Base        `bson:"job_base" json:"job_base" yaml:"job_base"`
	MigrationHelper `bson:"-" json:"-" yaml:"-"`
	mu              sync.Mutex
}

func (j *manualMigrationGenerator) Run() {
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
	j.mu.Lock()
	defer j.mu.Unlock()
	for iter.Next(&doc) {
		m := NewManualMigration(env, ManualMigration{
			ID:            doc.ID,
			OperationName: j.OperationName,
			Migration:     j.ID(),
			Namespace:     j.NS,
		}).(*manualMigrationJob)

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

func (j *simpleMigrationGenerator) Jobs() <-chan amboy.Job {
	env := j.Env()

	j.mu.Lock()
	defer j.mu.Unlock()

	out, err := generator(env, j.ID(), j.Migrations...)
	grip.CatchError(err)
	grip.Info("produced %d tasks for migration %s", len(j.Migrations), j.ID())
	j.Migrations = []*manualMigrationJob{}
	return out
}
