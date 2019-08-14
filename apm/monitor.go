package apm

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/event"
)

type Monitor struct {
	inProg     map[int64]eventKey
	inProgLock sync.Mutex

	current        map[eventKey]*eventRecord
	currentStartAt time.Time
	currentLock    sync.Mutex

	windows     []eventWindow
	windowsLock sync.Mutex
}

type eventKey struct {
	dbName   string
	cmdName  string
	collName string
}

type eventRecord struct {
	failCount    int64
	successCount int64
	durationTime time.Duration
	mutex        sync.Mutex
}

type eventWindow struct {
	timestamp time.Time
	data      map[eventKey]*eventRecord
}

func (m *Monitor) handleStartedEvent(ctx context.Context, e *event.CommandStartedEvent) {
	r := eventKey{
		dbName:  e.DatabaseName,
		cmdName: e.CommandName,
	}

	arg, err := e.Command.LookupErr(r.cmdName)
	if err == nil {
		r.collName, _ = arg.StringValueOK()
	}

	m.inProgLock.Lock()
	defer m.inProgLock.Unlock()

	m.inProg[e.RequestID] = r
}

func (m *Monitor) popRequest(id int64) eventKey {
	m.inProgLock.Lock()
	defer m.inProgLock.Unlock()

	out := m.inProg[id]
	delete(m.inProg, id)
	return out
}

func (m *Monitor) getRecord(id int64) *eventRecord {
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

func (m *Monitor) handleSuccessEvent(ctx context.Context, e *event.CommandSucceededEvent) {
	event := m.getRecord(e.RequestID)
	if event == nil {
		return
	}

	event.mutex.Lock()
	defer event.mutex.Unlock()

	event.successCount++
	event.durationTime += time.Duration(e.DurationNanos)
}

func (m *Monitor) handleFailedEvent(ctx context.Context, e *event.CommandFailedEvent) {
	event := m.getRecord(e.RequestID)
	if event == nil {
		return
	}

	event.mutex.Lock()
	defer event.mutex.Unlock()

	event.failCount++
	event.durationTime += time.Duration(e.DurationNanos)
}

func (m *Monitor) DriverAPM() event.CommandMonitor {
	return event.CommandMonitor{
		Started:   m.handleStartedEvent,
		Succeeded: m.handleSuccessEvent,
		Failed:    m.handleFailedEvent,
	}
}

func (m *Monitor) rotateCurrent() eventWindow {
	m.currentLock.Lock()
	defer m.currentLock.Unlock()

	out := eventWindow{
		data:      m.current,
		timestamp: m.currentStartAt,
	}
	m.current = make(map[eventKey]*eventRecord)
	m.currentStartAt = time.Now()
	return out
}

func (m *Monitor) Rotate() {
	m.windowsLock.Lock()
	defer m.windowsLock.Unlock()
	m.windows = append(m.windows, m.rotateCurrent())

	if len(m.windows) > 100 {
		m.windows = m.windows[1:]
	}
}
