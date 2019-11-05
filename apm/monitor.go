package apm

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/event"
)

type basicMonitor struct {
	config MonitorConfig

	inProg     map[int64]eventKey
	inProgLock sync.Mutex

	current        map[eventKey]*eventRecord
	currentStartAt time.Time
	currentLock    sync.Mutex
}

// NewwBasicMonitor returns a simple monitor implementation that does
// not automatically rotate data.
func NewBasicMonitor() Monitor {
	return &basicMonitor{
		inProg:  make(map[int64]eventKey),
		current: make(map[eventKey]*eventRecord),
	}
}

func (m *basicMonitor) popRequest(id int64) eventKey {
	m.inProgLock.Lock()
	defer m.inProgLock.Unlock()

	out := m.inProg[id]
	delete(m.inProg, id)
	return out
}

func (m *basicMonitor) setRequest(id int64, key eventKey) {
	m.inProgLock.Lock()
	defer m.inProgLock.Unlock()

	m.inProg[id] = key
}

func (m *basicMonitor) getRecord(id int64) *eventRecord {
	key := m.popRequest(id)
	if key.dbName == "" {
		return nil
	}

	m.currentLock.Lock()
	defer m.currentLock.Unlock()

	event := m.current[key]
	if event == nil {
		event = &eventRecord{}
		m.current[key] = event
	}

	return event
}

func (m *basicMonitor) DriverAPM() event.CommandMonitor {
	return event.CommandMonitor{
		Started: func(ctx context.Context, e *event.CommandStartedEvent) {
			var collName string

			if e.CommandName == "getMore" {
				collName, _ = e.Command.Lookup("collection").StringValueOK()
			} else {
				collName, _ = e.Command.Lookup(e.CommandName).StringValueOK()
			}

			m.setRequest(e.RequestID, eventKey{
				dbName:   e.DatabaseName,
				cmdName:  e.CommandName,
				collName: collName,
			})
		},
		Succeeded: func(ctx context.Context, e *event.CommandSucceededEvent) {
			event := m.getRecord(e.RequestID)
			if event == nil {
				return
			}

			event.mutex.Lock()
			defer event.mutex.Unlock()

			event.Succeeded++
			event.Duration += time.Duration(e.DurationNanos)
		},
		Failed: func(ctx context.Context, e *event.CommandFailedEvent) {
			event := m.getRecord(e.RequestID)
			if event == nil {
				return
			}

			event.mutex.Lock()
			defer event.mutex.Unlock()

			event.Failed++
			event.Duration += time.Duration(e.DurationNanos)
		},
	}
}

func (m *basicMonitor) Rotate() Event {
	m.currentLock.Lock()
	defer m.currentLock.Unlock()

	out := &eventWindow{
		data:      m.current,
		timestamp: m.currentStartAt,
	}
	m.current = make(map[eventKey]*eventRecord)
	m.currentStartAt = time.Now()
	return out
}
