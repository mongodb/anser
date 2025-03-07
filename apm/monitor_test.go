package apm

import (
	"context"
	"testing"
	"time"

	"github.com/evergreen-ci/birch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
)

func TestMonitor(t *testing.T) {
	m, ok := NewBasicMonitor(nil).(*basicMonitor)
	require.True(t, ok)
	t.Run("Tracking", func(t *testing.T) {
		t.Run("Pop", func(t *testing.T) {
			t.Run("Empty", func(t *testing.T) {
				assert.Len(t, m.inProg, 0)
				key := m.popRequest(42)
				assert.Zero(t, key)
				assert.Len(t, m.inProg, 0)
			})
			t.Run("Existing", func(t *testing.T) {
				m.inProg[42] = eventKey{dbName: "amboy"}
				key := m.popRequest(42)
				assert.Equal(t, "amboy", key.dbName)
				assert.Len(t, m.inProg, 0)
			})
		})
		t.Run("Set", func(t *testing.T) {
			resetMonitor(t, m)
			t.Run("WithValue", func(t *testing.T) {
				m.setRequest(42, eventKey{cmdName: "find"})
				assert.Len(t, m.inProg, 1)
				k := m.popRequest(42)
				assert.Equal(t, "find", k.cmdName)
			})
			t.Run("Filter", func(t *testing.T) {
				m.config = &MonitorConfig{
					Databases: []string{"amboy"},
					Commands:  []string{"find"},
				}
				assert.Len(t, m.inProg, 0)
				m.setRequest(42, eventKey{cmdName: "find"})
				assert.Len(t, m.inProg, 0)
				m.setRequest(42, eventKey{dbName: "amboy", cmdName: "find"})
				assert.Len(t, m.inProg, 1)
				m.config = nil
			})
		})
		t.Run("Get", func(t *testing.T) {
			resetMonitor(t, m)
			t.Run("Empty", func(t *testing.T) {
				r := m.getRecord(42)
				require.Nil(t, r)

				assert.Len(t, m.current, 0)
			})
			t.Run("Zeroed", func(t *testing.T) {
				resetMonitor(t, m)
				m.inProg[42] = eventKey{}
				r := m.getRecord(42)
				assert.Len(t, m.inProg, 0)
				require.Nil(t, r)

				assert.Len(t, m.current, 0)
			})
			t.Run("PartialData", func(t *testing.T) {
				resetMonitor(t, m)
				m.inProg[42] = eventKey{dbName: "amboy", cmdName: "find"}
				r := m.getRecord(42)
				assert.NotNil(t, r)
				assert.Len(t, m.inProg, 0)

				assert.Len(t, m.current, 1)
			})
			t.Run("MultipleData", func(t *testing.T) {
				resetMonitor(t, m)
				m.inProg[42] = eventKey{dbName: "amboy", collName: "jobs", cmdName: "find"}
				r := m.getRecord(42)
				assert.NotNil(t, r)
				assert.Len(t, m.inProg, 0)

				assert.Len(t, m.current, 1)
			})
		})
	})
	t.Run("Collector", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		resetMonitor(t, m)
		collector := m.DriverAPM()
		require.NotNil(t, collector)
		t.Run("StartEventFind", func(t *testing.T) {
			resetMonitor(t, m)
			assert.Len(t, m.inProg, 0)
			collector.Started(ctx, &event.CommandStartedEvent{
				DatabaseName: "amboy",
				CommandName:  "find",
				RequestID:    42,
				Command: buildCommand(t, birch.DC.Elements(
					birch.EC.String("find", "jobs"),
				)),
			})
			assert.Len(t, m.inProg, 1)
		})
		t.Run("StartEventGetMore", func(t *testing.T) {
			resetMonitor(t, m)
			collector.Started(ctx, &event.CommandStartedEvent{
				DatabaseName: "amboy",
				CommandName:  "getMore",
				RequestID:    44,
				Command: buildCommand(t, birch.DC.Elements(
					birch.EC.String("getMore", ""),
					birch.EC.String("collection", "jobs"),
				)),
			})
			assert.Equal(t, "jobs", m.inProg[44].collName)
		})
		t.Run("InvalidCommand", func(t *testing.T) {
			resetMonitor(t, m)
			collector.Started(ctx, &event.CommandStartedEvent{
				DatabaseName: "amboy",
				CommandName:  "wat",
				RequestID:    84,
				Command:      nil,
			})
			_, ok = m.inProg[84]
			assert.True(t, ok)
			assert.Equal(t, "", m.inProg[84].collName)
			assert.Equal(t, "wat", m.inProg[84].cmdName)
		})
		t.Run("CompleteNils", func(t *testing.T) {
			resetMonitor(t, m)
			collector.Succeeded(ctx, &event.CommandSucceededEvent{CommandFinishedEvent: event.CommandFinishedEvent{RequestID: 100}})
			assert.Len(t, m.current, 0)
			collector.Failed(ctx, &event.CommandFailedEvent{CommandFinishedEvent: event.CommandFinishedEvent{RequestID: 100}})
			assert.Len(t, m.current, 0)
		})
		t.Run("Success", func(t *testing.T) {
			resetMonitor(t, m)
			collector.Started(ctx, &event.CommandStartedEvent{
				DatabaseName: "amboy",
				CommandName:  "find",
				RequestID:    42,
				Command: buildCommand(t, birch.DC.Elements(
					birch.EC.String("find", "jobs"),
				)),
			})

			collector.Succeeded(ctx, &event.CommandSucceededEvent{CommandFinishedEvent: event.CommandFinishedEvent{RequestID: 42}})
			assert.Len(t, m.current, 1)

			var op *eventRecord
			op, ok = m.current[eventKey{dbName: "amboy", cmdName: "find", collName: "jobs"}]
			require.True(t, ok)
			assert.EqualValues(t, 1, op.Succeeded)
			assert.EqualValues(t, 0, op.Failed)
		})
		t.Run("Failed", func(t *testing.T) {
			resetMonitor(t, m)
			collector.Started(ctx, &event.CommandStartedEvent{
				DatabaseName: "amboy",
				CommandName:  "aggregate",
				RequestID:    100,
				Command: buildCommand(t, birch.DC.Elements(
					birch.EC.String("aggregate", "group.jobs"),
				)),
			})

			collector.Failed(ctx, &event.CommandFailedEvent{CommandFinishedEvent: event.CommandFinishedEvent{RequestID: 100}})
			assert.Len(t, m.current, 1)

			var op *eventRecord
			op, ok = m.current[eventKey{dbName: "amboy", cmdName: "aggregate", collName: "group.jobs"}]
			require.True(t, ok)
			assert.EqualValues(t, 0, op.Succeeded)
			assert.EqualValues(t, 1, op.Failed)
		})
		t.Run("Wrapper", func(t *testing.T) {
			m, ok = NewBasicMonitor(nil).(*basicMonitor)
			require.True(t, ok)
			nctx, ncancel := context.WithCancel(ctx)
			defer ncancel()
			wrapped := NewLoggingMonitor(nctx, 10*time.Millisecond, m)
			assert.NotNil(t, wrapped)
			assert.Implements(t, (*Monitor)(nil), wrapped)
			time.Sleep(100 * time.Millisecond)
			resetMonitor(t, m)
		})
	})
	t.Run("Rotate", func(t *testing.T) {
		t.Run("Timestamp", func(t *testing.T) {
			m.currentLock.Lock() // We have to lock because race detector complains
			startedAt := m.currentStartAt
			m.currentLock.Unlock()
			_ = m.Rotate()
			assert.True(t, startedAt.Before(m.currentStartAt))
		})
		t.Run("Operation", func(t *testing.T) {
			assert.Len(t, m.current, 0)
			m.inProg[42] = eventKey{cmdName: "find"}
			_ = m.getRecord(42)
			assert.Len(t, m.current, 1)
			e := m.Rotate()
			assert.Len(t, m.current, 0)

			if event, ok := e.(*eventWindow); ok {
				assert.Equal(t, 1, len(event.data))
			}
		})
		t.Run("PropagatesSettings", func(t *testing.T) {
			conf := &MonitorConfig{AllTags: true, Tags: []string{"41", "1"}}
			monitor := NewBasicMonitor(conf)
			event, ok := monitor.Rotate().(*eventWindow)
			require.True(t, ok)
			assert.True(t, event.allTags)
			assert.Contains(t, event.tags, "41")
			assert.Contains(t, event.tags, "1")
		})
	})
	t.Run("Tags", func(t *testing.T) {
		conf := &MonitorConfig{Tags: []string{"41", "1"}}
		t.Run("ConstructorSorts", func(t *testing.T) {
			assert.Equal(t, "41", conf.Tags[0])
			assert.Equal(t, "1", conf.Tags[1])

			monitor := NewBasicMonitor(conf)
			require.NotNil(t, monitor)
			assert.Equal(t, "1", conf.Tags[0])
			assert.Equal(t, "41", conf.Tags[1])
		})
		t.Run("AddTags", func(t *testing.T) {
			ctx := SetTags(context.Background(), "41")
			monitor := NewBasicMonitor(conf).(*basicMonitor)
			event := &eventRecord{Tags: map[string]int64{}}
			monitor.addTags(ctx, event)
			assert.Contains(t, event.Tags, "41")
		})
		t.Run("AddTagsSkipped", func(t *testing.T) {
			ctx := SetTags(context.Background(), "400")
			monitor := NewBasicMonitor(conf).(*basicMonitor)
			event := &eventRecord{Tags: map[string]int64{}}
			monitor.addTags(ctx, event)
			assert.Len(t, event.Tags, 0)
		})
	})

}

func resetMonitor(t *testing.T, in Monitor) {
	switch m := in.(type) {
	case *basicMonitor:
		m.currentLock.Lock()
		defer m.currentLock.Unlock()
		m.current = m.config.window()
		require.Len(t, m.current, 0)

		m.inProgLock.Lock()
		defer m.inProgLock.Unlock()
		m.inProg = map[int64]eventKey{}

	case *loggingMonitor:
		resetMonitor(t, m.Monitor)
	}
}

func buildCommand(t *testing.T, doc *birch.Document) bson.Raw {
	raw, err := doc.MarshalBSON()
	require.NoError(t, err)
	return bson.Raw(raw)
}
