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
		Generators: []MigrationDefinition{
			{
				Name: "first",
				Job: NewSimpleMigrationGenerator(
					// environment and setup:
					env, ns,
					// query:
					map[string]interface{}{
						"time": map[string]interface{}{"$gt": time.Now().Add(-time.Hour)},
					},
					// update:
					map[string]interface{}{
						"$rename": map[string]string{"time": "timeSince"},
					}),
				DependsOn: []string{},
			},
			{
				Name: "second",
			},
		},
	}

	app.Setup(env)
}
