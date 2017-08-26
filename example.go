package anser

import "time"

////////////////////////////////////////////////////////////////////////
//
// mostly as an example...
//
func main() {
	env := GetEnvironment()
	ns := Namespace{DB: "mci", Collection: "test"}

	app := MigrationApplication{
		Generators: []MigrationGenerator{
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
			NewAggregateMigrationGenerator(env,
				GeneratorOptions{
					JobID:     "second",
					DependsOn: []string{"first"},
					NS:        ns,
					Query: map[string]interface{}{
						"time": map[string]interface{}{"$gt": time.Now().Add(-time.Hour)},
					},
				},
				"op-name"),
		},
	}

	app.Setup(env)
}
