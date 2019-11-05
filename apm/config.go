package apm

type MonitorConf struct {
	PopulateEvents bool
	Commands       []string
	Databases      []string
	Collections    []string
}

func stringSliceContains(slice []string, item string) bool {
	if len(slice) == 0 {
		return false
	}

	for idx := range slice {
		if slice[idx] == item {
			return true
		}
	}

	return false
}

func (c *MonitorConf) shouldTrack(e eventKey) bool {
	if c == nil {
		return true
	}

	if len(c.Databases) > 0 && !stringSliceContains(c.Databases, e.dbName) {
		return false
	}

	if len(c.Collections) > 0 && !stringSliceContains(c.Collections, e.collName) {
		return false
	}

	if len(c.Commands) > 0 && !stringSliceContains(c.Commands, e.cmdName) {
		return false
	}

	return true
}

func (c *MonitorConf) window() map[eventKey]*eventRecord {
	out := make(map[eventKey]*eventRecord)
	if c == nil {
		return out
	}

	if !c.PopulateEvents {
		return out
	}

	for _, db := range c.Databases {
		for _, coll := range c.Collections {
			for _, cmd := range c.Commands {
				out[eventKey{dbName: db, collName: coll, cmdName: cmd}] = &eventRecord{}
			}
		}
	}

	return out
}
