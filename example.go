package anser

import (
	"time"

	"github.com/mongodb/grip"
	"github.com/tychoish/anser/model"
	"golang.org/x/net/context"
)

////////////////////////////////////////////////////////////////////////
//
// mostly as an example...
//
func main() {
	env := GetEnvironment()
	ns := model.Namespace{DB: "mci", Collection: "test"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &Application{
		Generators: []Generator{
			NewSimpleMigrationGenerator(env,
				GeneratorOptions{
					JobID:     "first",
					DependsOn: []string{},
					NS:        ns,
					Query: map[string]interface{}{
						"time": map[string]interface{}{"$gt": time.Now().Add(-time.Hour)},
					},
				},
				// update:
				map[string]interface{}{
					"$rename": map[string]string{"time": "timeSince"},
				}),
			NewStreamMigrationGenerator(env,
				GeneratorOptions{
					JobID:     "second",
					DependsOn: []string{"first"},
					NS:        ns,
					Query: map[string]interface{}{
						"time": map[string]interface{}{"$gt": time.Now().Add(-time.Hour)},
					},
				},
				// the name of a registered aggregate operation
				"op-name"),
		},
	}

	app.Setup(env)
	grip.CatchEmergencyFatal(app.Run(ctx))
}
