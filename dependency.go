package anser

import (
	"github.com/mongodb/amboy/dependency"
	"github.com/mongodb/amboy/registry"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

type migrationDependency struct {
	MigrationID string                 `bson:"migration" json:"migration" yaml:"migration"`
	Query       map[string]interface{} `bson:"query" json:"query" yaml:"query"`
	NS          Namespace              `bson:"namespace" json:"namespace" yaml:"namespace"`
	T           dependency.TypeInfo    `bson:"type" json:"type" yaml:"type"`

	MigrationHelper `bson:"-" json:"-" yaml:"-"`
	*dependency.JobEdges
}

func init() {
	registry.AddDependencyType("anser-migration", func() dependency.Manager { return makeMigrationDependencyManager() })
}

func makeMigrationDependencyManager() *migrationDependency {
	return &migrationDependency{
		JobEdges: dependency.NewJobEdges(),
		T: dependency.TypeInfo{
			Name:    "anser-migration",
			Version: 0,
		},
	}
}

// NewMigrationDependencyManager constructs a new
func NewMigrationDependencyManager(e Environment, id string,
	q map[string]interface{}, ns Namespace) (dependency.Manager, error) {

	d := makeMigrationDependencyManager()
	if err := d.SetEnv(e); err != nil {
		return nil, errors.Wrap(err, "problem with environment")
	}
	d.Query = q
	d.MigrationID = id

	return d, nil
}

func (d *migrationDependency) Type() dependency.TypeInfo {
	return d.T
}

func (d *migrationDependency) State() dependency.State {
	env := d.Env()

	session, err := env.GetSession()
	if err != nil {
		grip.Error(err)
		return dependency.Unresolved
	}
	defer session.Close()

	coll := session.DB(d.NS.DB).C(d.NS.Collection)

	num, err := coll.Find(bson.M(d.Query)).Count()
	if err != nil {
		grip.Error(err)
		return dependency.Unresolved
	}

	if num == 0 {
		return dependency.Passed
	}

	edges := d.Edges()
	if len(edges) == 0 {
		return dependency.Ready
	}

	// query the "done" dependencies, and make sure that all the
	// edges listed in the edges document are satisfied.
	ns := env.MetadataNamespace()

	count := 0
	meta := &MigrationMetadata{}
	iter := session.DB(ns.DB).C(ns.Collection).Find(getDependencyStateQuery(edges)).Iter()
	for iter.Next(meta) {
		if !meta.Satisfied() {
			return dependency.Blocked
		}
		count++
	}

	if err := iter.Close(); err != nil {
		grip.Warning(err)
		return dependency.Blocked
	}

	if count < len(edges) {
		return dependency.Blocked
	}

	return dependency.Ready
}

func getDependencyStateQuery(ids []string) bson.M {
	orQuery := make([]bson.M, len(ids))

	for idx, id := range ids {
		orQuery[idx] = bson.M{"_id": id}
	}

	return bson.M{"$or": orQuery}
}
