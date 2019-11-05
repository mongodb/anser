package apm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.Len(t, m.current, 0)
			t.Run("Empty", func(t *testing.T) {
				r := m.getRecord(42)
				require.NotNil(t, r)

				assert.Len(t, m.current, 1)
			})
			t.Run("Zeroed", func(t *testing.T) {
				m.inProg[42] = eventKey{}
				r := m.getRecord(42)
				assert.Len(t, m.inProg, 0)
				require.NotNil(t, r)

				assert.Len(t, m.current, 2)
			})
			t.Run("PartialData", func(t *testing.T) {
				m.inProg[42] = eventKey{dbName: "amboy", cmdName: "find"}
				r := m.getRecord(42)
				assert.NotNil(t, r)
				assert.Len(t, m.inProg, 0)

				assert.Len(t, m.current, 2)
			})
			t.Run("MultipleData", func(t *testing.T) {
				m.inProg[42] = eventKey{dbName: "amboy", collName: "jobs", cmdName: "find"}
				r := m.getRecord(42)
				assert.NotNil(t, r)
				assert.Len(t, m.inProg, 0)

				assert.Len(t, m.current, 3)
			})
			m.current = m.config.window()
			assert.Len(t, m.current, 0)
		})
	})
	t.Run("Collector", func(t *testing.T) {
		assert.NotNil(t, m.DriverAPM())
	})
	t.Run("Rotate", func(t *testing.T) {
		t.Run("Timestamp", func(t *testing.T) {
			startedAt := m.currentStartAt
			_ = m.Rotate()
			assert.True(t, startedAt.Before(m.currentStartAt))
		})
		t.Run("Rotate", func(t *testing.T) {
			assert.Len(t, m.current, 0)
			_ = m.getRecord(42)
			assert.Len(t, m.current, 1)
			e := m.Rotate()
			assert.Len(t, m.current, 0)

			if event, ok := e.(*eventWindow); ok {
				assert.Equal(t, 1, len(event.data))
			}
		})
	})
}
