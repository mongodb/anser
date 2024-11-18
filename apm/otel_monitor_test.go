package apm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestCompactArray(t *testing.T) {
	_, type1, err := bson.MarshalValue("type 1")
	require.NoError(t, err)
	_, type2, err := bson.MarshalValue("type 2")
	require.NoError(t, err)

	for name, testCase := range map[string]struct {
		input     bson.A
		unchanged bool
		expected  bson.A
	}{
		"emptyArray": {input: bson.A{}, unchanged: true},
		"corruptValue": {
			input: bson.A{
				bson.RawValue{Type: bson.TypeString, Value: []byte("invalid bson")},
				bson.RawValue{Type: bson.TypeString, Value: type1},
			},
			unchanged: true,
		},
		"arrayType": {
			input: bson.A{
				bson.RawValue{Type: bson.TypeString, Value: type1},
				bson.RawValue{Type: bson.TypeArray},
			},
			unchanged: true,
		},
		"documentType": {
			input: bson.A{
				bson.RawValue{Type: bson.TypeString, Value: type1},
				bson.RawValue{Type: bson.TypeEmbeddedDocument},
			},
			unchanged: true,
		},
		"multiplesOfEachType": {
			input: bson.A{
				bson.RawValue{Type: bson.TypeString, Value: type1},
				bson.RawValue{Type: bson.TypeString, Value: type1},
				bson.RawValue{Type: bson.TypeString, Value: type2},
				bson.RawValue{Type: bson.TypeString, Value: type2},
			},
			expected: bson.A{
				bson.RawValue{Type: bson.TypeString, Value: type1},
				bson.RawValue{Type: bson.TypeString, Value: type2},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if testCase.unchanged {
				assert.Equal(t, testCase.input, compactArray(testCase.input))
			} else {
				assert.Equal(t, testCase.expected, compactArray(testCase.input))
			}
		})
	}
}

func TestStripDocument(t *testing.T) {
	for name, testCase := range map[string]struct {
		input       bson.D
		errExpected bool
		expected    bson.D
	}{
		"simpleValues": {
			input: bson.D{
				{Key: "a_string", Value: "Why did the computer go broke? Because it used up all its cache!"},
				{Key: "an_int", Value: 1},
				{Key: "a_date", Value: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
			},
			expected: bson.D{
				{Key: "a_string", Value: "<string>"},
				{Key: "an_int", Value: "<32-bit integer>"},
				{Key: "a_date", Value: "<UTC datetime>"},
			},
		},
		"nestedArray": {
			input: bson.D{
				{Key: "array", Value: []interface{}{"one", 2, "two", 3, time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)}},
			},
			expected: bson.D{
				{Key: "array", Value: []interface{}{"<string>", "<32-bit integer>", "<UTC datetime>"}},
			},
		},
		"nestedSubdocument": {
			input: bson.D{
				{Key: "subdocument", Value: bson.M{"my_int": 1}},
			},
			expected: bson.D{
				{Key: "subdocument", Value: bson.M{"my_int": "<32-bit integer>"}},
			},
		},
		"nestedRecursively": {
			input: bson.D{
				{Key: "subdocument", Value: bson.M{"array": []interface{}{"one"}}},
			},
			expected: bson.D{
				{Key: "subdocument", Value: bson.M{"array": []interface{}{"<string>"}}},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			input, err := bson.Marshal(testCase.input)
			require.NoError(t, err)
			expectedOutput, err := bson.MarshalExtJSON(testCase.expected, false, false)
			require.NoError(t, err)

			if testCase.errExpected {
				_, err := stripDocument(input)
				assert.Error(t, err)
			} else {
				val, err := stripDocument(input)
				assert.NoError(t, err)
				valString, err := bson.MarshalExtJSON(val, false, false)
				assert.NoError(t, err)
				assert.Equal(t, expectedOutput, valString)
			}
		})
	}
}

func TestExtractStatement(t *testing.T) {
	for name, testCase := range map[string]struct {
		input       string
		commandName string
		stripped    bool
		expected    string
	}{
		"aggregate": {
			commandName: "aggregate",
			input:       `{"aggregate":"evg.service.group","pipeline":[{"$match":{"group":"service.host.termination"}},{"$group":{"_id":1,"n":{"$sum":1}}}],"cursor":{},"readConcern":{"level":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731958801,"i":16}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299999,"$db":"amboy"}`,
			expected: `{
  "pipeline": [
    {
      "$match": {
        "group": "service.host.termination"
      }
    },
    {
      "$group": {
        "_id": 1,
        "n": {
          "$sum": 1
        }
      }
    }
  ]
}`,
			stripped: false,
		},
		"aggregateStripped": {
			commandName: "aggregate",
			input:       `{"aggregate":"evg.service.group","pipeline":[{"$match":{"group":"service.host.termination"}},{"$group":{"_id":1,"n":{"$sum":1}}}],"cursor":{},"readConcern":{"level":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731958801,"i":16}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299999,"$db":"amboy"}`,
			expected: `{
  "pipeline": [
    {
      "$match": {
        "group": "<string>"
      }
    },
    {
      "$group": {
        "_id": "<32-bit integer>",
        "n": {
          "$sum": "<32-bit integer>"
        }
      }
    }
  ]
}`,
			stripped: true,
		},
		"find": {
			commandName: "find",
			input:       `{"find":"admin","filter":{"_id":{"$in":["service_flags","tracer","pod_lifecycle","sleep_schedule","api","cedar","commit_queue","host_jasper","amboy_db","project_creation","spawnhost","github_check_run","slack","ui","hostinit","jira","notify","scheduler","amboy","providers","repotracker","auth","jira_notifications","task_limits","triggers","runtime_environments","splunk","global","buckets","container_pools","logger_config","newrelic"]}},"readConcern":{"level":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731961217,"i":520}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"$db":"mci"}`,
			expected: `{
  "filter": {
    "_id": {
      "$in": [
        "service_flags",
        "tracer",
        "pod_lifecycle",
        "sleep_schedule",
        "api",
        "cedar",
        "commit_queue",
        "host_jasper",
        "amboy_db",
        "project_creation",
        "spawnhost",
        "github_check_run",
        "slack",
        "ui",
        "hostinit",
        "jira",
        "notify",
        "scheduler",
        "amboy",
        "providers",
        "repotracker",
        "auth",
        "jira_notifications",
        "task_limits",
        "triggers",
        "runtime_environments",
        "splunk",
        "global",
        "buckets",
        "container_pools",
        "logger_config",
        "newrelic"
      ]
    }
  }
}`,
		},
		"findStripped": {
			commandName: "find",
			input:       `{"find":"admin","filter":{"_id":{"$in":["service_flags","tracer","pod_lifecycle","sleep_schedule","api","cedar","commit_queue","host_jasper","amboy_db","project_creation","spawnhost","github_check_run","slack","ui","hostinit","jira","notify","scheduler","amboy","providers","repotracker","auth","jira_notifications","task_limits","triggers","runtime_environments","splunk","global","buckets","container_pools","logger_config","newrelic"]}},"readConcern":{"level":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731961217,"i":520}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"$db":"mci"}`,
			expected: `{
  "filter": {
    "_id": {
      "$in": [
        "<string>"
      ]
    }
  }
}`,
			stripped: true,
		},
		"update": {
			commandName: "update",
			input:       `{"update":"tasks","ordered":true,"writeConcern":{"w":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731958835,"i":578}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299994,"$db":"mci","updates":[{"q":{"activated_time":{"$lte":{"$date":"2024-11-11T19:40:35.484Z"}},"activated":true,"status":"undispatched","priority":{"$gt":-1},"$and":[{"$or":[{"execution_platform":{"$exists":false}},{"execution_platform":"host"}]},{"$or":[{"unattainable_dependency":false},{"override_dependencies":true}]}],"distro":{"$in":["ubuntu1804-small","ubuntu1804","ubuntu1804-test"]}},"u":{"$set":{"priority":-1,"activated":false}},"multi":true,"hint":{"distro":1,"status":1,"activated":1,"priority":1,"override_dependencies":1,"unattainable_dependency":1}}]}`,
			expected: `{
  "q": {
    "activated_time": {
      "$lte": {
        "$date": "2024-11-11T19:40:35.484Z"
      }
    },
    "activated": true,
    "status": "undispatched",
    "priority": {
      "$gt": -1
    },
    "$and": [
      {
        "$or": [
          {
            "execution_platform": {
              "$exists": false
            }
          },
          {
            "execution_platform": "host"
          }
        ]
      },
      {
        "$or": [
          {
            "unattainable_dependency": false
          },
          {
            "override_dependencies": true
          }
        ]
      }
    ],
    "distro": {
      "$in": [
        "ubuntu1804-small",
        "ubuntu1804",
        "ubuntu1804-test"
      ]
    }
  },
  "u": {
    "$set": {
      "priority": -1,
      "activated": false
    }
  },
  "multi": true,
  "hint": {
    "distro": 1,
    "status": 1,
    "activated": 1,
    "priority": 1,
    "override_dependencies": 1,
    "unattainable_dependency": 1
  }
}`,
		},
		"updateStripped": {
			commandName: "update",
			input:       `{"update":"tasks","ordered":true,"writeConcern":{"w":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731958835,"i":578}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299994,"$db":"mci","updates":[{"q":{"activated_time":{"$lte":{"$date":"2024-11-11T19:40:35.484Z"}},"activated":true,"status":"undispatched","priority":{"$gt":-1},"$and":[{"$or":[{"execution_platform":{"$exists":false}},{"execution_platform":"host"}]},{"$or":[{"unattainable_dependency":false},{"override_dependencies":true}]}],"distro":{"$in":["ubuntu1804-small","ubuntu1804","ubuntu1804-test"]}},"u":{"$set":{"priority":-1,"activated":false}},"multi":true,"hint":{"distro":1,"status":1,"activated":1,"priority":1,"override_dependencies":1,"unattainable_dependency":1}}]}`,
			expected: `{
  "q": {
    "activated_time": {
      "$lte": "<UTC datetime>"
    },
    "activated": "<boolean>",
    "status": "<string>",
    "priority": {
      "$gt": "<32-bit integer>"
    },
    "$and": [
      {
        "$or": [
          {
            "execution_platform": {
              "$exists": "<boolean>"
            }
          },
          {
            "execution_platform": "<string>"
          }
        ]
      },
      {
        "$or": [
          {
            "unattainable_dependency": "<boolean>"
          },
          {
            "override_dependencies": "<boolean>"
          }
        ]
      }
    ],
    "distro": {
      "$in": [
        "<string>"
      ]
    }
  },
  "u": {
    "$set": {
      "priority": "<32-bit integer>",
      "activated": "<boolean>"
    }
  },
  "multi": "<boolean>",
  "hint": {
    "distro": "<32-bit integer>",
    "status": "<32-bit integer>",
    "activated": "<32-bit integer>",
    "priority": "<32-bit integer>",
    "override_dependencies": "<32-bit integer>",
    "unattainable_dependency": "<32-bit integer>"
  }
}`,
			stripped: true,
		},
		"delete": {
			commandName: "delete",
			input:       `{"delete":"data_cache","ordered":true,"writeConcern":{"w":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"txnNumber":361,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731961014,"i":905}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299999,"$db":"mci","deletes":[{"q":{"_id":"https://api.github.com/repos/evergreen-ci/evergreen/commits/86445d84a490b1df45184c48160c66f4549e879a"},"limit":1}]}`,
			expected: `{
  "q": {
    "_id": "https://api.github.com/repos/evergreen-ci/evergreen/commits/86445d84a490b1df45184c48160c66f4549e879a"
  },
  "limit": 1
}`,
		},
		"deleteStripped": {
			commandName: "delete",
			input:       `{"delete":"data_cache","ordered":true,"writeConcern":{"w":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"txnNumber":361,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731961014,"i":905}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299999,"$db":"mci","deletes":[{"q":{"_id":"https://api.github.com/repos/evergreen-ci/evergreen/commits/86445d84a490b1df45184c48160c66f4549e879a"},"limit":1}]}`,
			expected: `{
  "q": {
    "_id": "<string>"
  },
  "limit": "<32-bit integer>"
}`,
			stripped: true,
		},
		"findAndModify": {
			commandName: "findAndModify",
			input:       `{"findAndModify":"hosts","new":true,"query":{"_id":"i-012345678901"},"update":{"$inc":{"total_idle_time":103516332212}},"upsert":false,"writeConcern":{"w":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"txnNumber":7795,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731960233,"i":1022}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299999,"$db":"mci"}`,
			expected: `{
  "query": {
    "_id": "i-012345678901"
  },
  "update": {
    "$inc": {
      "total_idle_time": 103516332212
    }
  },
  "upsert": false
}`,
		},
		"findAndModifyStripped": {
			commandName: "findAndModify",
			input:       `{"findAndModify":"hosts","new":true,"query":{"_id":"i-012345678901"},"update":{"$inc":{"total_idle_time":103516332212}},"upsert":false,"writeConcern":{"w":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"txnNumber":7795,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731960233,"i":1022}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299999,"$db":"mci"}`,
			expected: `{
  "query": {
    "_id": "<string>"
  },
  "update": {
    "$inc": {
      "total_idle_time": "<64-bit integer>"
    }
  },
  "upsert": "<boolean>"
}`,
			stripped: true,
		},
		"insert": {
			commandName: "insert",
			input:       `{"insert":"evg.service.group","ordered":false,"writeConcern":{"w":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"txnNumber":1192,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731959774,"i":1}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299999,"$db":"amboy","documents":[{"_id":"service.generate.tasks.version.12345.generate-tasks-example_task_id","type":"generate-tasks","group":"service.generate.tasks.version.12345","version":0,"priority":0,"status":{"owner":"","completed":false,"in_prog":false,"mod_ts":{"$date":{"$numberLong":"-62135596800000"}},"mod_count":0,"err_count":0},"scopes":["generate-tasks.example_task_id"],"enqueue_scopes":["generate-tasks.example_task_id"],"retry_info":{"retryable":false,"needs_retry":false,"current_attempt":0},"time_info":{"created":{"$date":"2024-11-18T19:56:14.733Z"}},"job":{"job_base":{"name":"generate-tasks-example_task_id","job_type":{"name":"generate-tasks","version":0},"required_scopes":["generate-tasks.12345","generate-tasks.example_task_id"]},"task_id":"example_task_id"},"dependency":{"type":"always","version":0,"edges":[],"dependency":{"type":{"name":"always","version":0},"jobedges":{"edges":[]}}}}]}`,
			expected: `{
  "ordered": false,
  "documents": [
    {
      "_id": "service.generate.tasks.version.12345.generate-tasks-example_task_id",
      "type": "generate-tasks",
      "group": "service.generate.tasks.version.12345",
      "version": 0,
      "priority": 0,
      "status": {
        "owner": "",
        "completed": false,
        "in_prog": false,
        "mod_ts": {
          "$date": {
            "$numberLong": "-62135596800000"
          }
        },
        "mod_count": 0,
        "err_count": 0
      },
      "scopes": [
        "generate-tasks.example_task_id"
      ],
      "enqueue_scopes": [
        "generate-tasks.example_task_id"
      ],
      "retry_info": {
        "retryable": false,
        "needs_retry": false,
        "current_attempt": 0
      },
      "time_info": {
        "created": {
          "$date": "2024-11-18T19:56:14.733Z"
        }
      },
      "job": {
        "job_base": {
          "name": "generate-tasks-example_task_id",
          "job_type": {
            "name": "generate-tasks",
            "version": 0
          },
          "required_scopes": [
            "generate-tasks.12345",
            "generate-tasks.example_task_id"
          ]
        },
        "task_id": "example_task_id"
      },
      "dependency": {
        "type": "always",
        "version": 0,
        "edges": [],
        "dependency": {
          "type": {
            "name": "always",
            "version": 0
          },
          "jobedges": {
            "edges": []
          }
        }
      }
    }
  ]
}`,
		},
		"insertStripped": {
			commandName: "insert",
			input:       `{"insert":"evg.service.group","ordered":false,"writeConcern":{"w":"majority"},"lsid":{"id":{"$binary":{"base64":"aggafjahdfLSJKDF3as5Fg==","subType":"04"}}},"txnNumber":1192,"$clusterTime":{"clusterTime":{"$timestamp":{"t":1731959774,"i":1}},"signature":{"hash":{"$binary":{"base64":"asdksdh2afksfas+Fas/djasdEo=","subType":"00"}},"keyId":1234567890987654321}},"maxTimeMS":299999,"$db":"amboy","documents":[{"_id":"service.generate.tasks.version.12345.generate-tasks-example_task_id","type":"generate-tasks","group":"service.generate.tasks.version.12345","version":0,"priority":0,"status":{"owner":"","completed":false,"in_prog":false,"mod_ts":{"$date":{"$numberLong":"-62135596800000"}},"mod_count":0,"err_count":0},"scopes":["generate-tasks.example_task_id"],"enqueue_scopes":["generate-tasks.example_task_id"],"retry_info":{"retryable":false,"needs_retry":false,"current_attempt":0},"time_info":{"created":{"$date":"2024-11-18T19:56:14.733Z"}},"job":{"job_base":{"name":"generate-tasks-example_task_id","job_type":{"name":"generate-tasks","version":0},"required_scopes":["generate-tasks.12345","generate-tasks.example_task_id"]},"task_id":"example_task_id"},"dependency":{"type":"always","version":0,"edges":[],"dependency":{"type":{"name":"always","version":0},"jobedges":{"edges":[]}}}}]}`,
			expected: `{
  "ordered": "<boolean>",
  "documents": [
    {
      "_id": "<string>",
      "type": "<string>",
      "group": "<string>",
      "version": "<32-bit integer>",
      "priority": "<32-bit integer>",
      "status": {
        "owner": "<string>",
        "completed": "<boolean>",
        "in_prog": "<boolean>",
        "mod_ts": "<UTC datetime>",
        "mod_count": "<32-bit integer>",
        "err_count": "<32-bit integer>"
      },
      "scopes": [
        "<string>"
      ],
      "enqueue_scopes": [
        "<string>"
      ],
      "retry_info": {
        "retryable": "<boolean>",
        "needs_retry": "<boolean>",
        "current_attempt": "<32-bit integer>"
      },
      "time_info": {
        "created": "<UTC datetime>"
      },
      "job": {
        "job_base": {
          "name": "<string>",
          "job_type": {
            "name": "<string>",
            "version": "<32-bit integer>"
          },
          "required_scopes": [
            "<string>"
          ]
        },
        "task_id": "<string>"
      },
      "dependency": {
        "type": "<string>",
        "version": "<32-bit integer>",
        "edges": [],
        "dependency": {
          "type": {
            "name": "<string>",
            "version": "<32-bit integer>"
          },
          "jobedges": {
            "edges": []
          }
        }
      }
    }
  ]
}`,
			stripped: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			val, err := extractStatement(testCase.commandName, testCase.input, testCase.stripped)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, val)
		})
	}
}
