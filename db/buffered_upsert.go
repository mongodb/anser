package db

import (
	"context"
	"time"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

type anserBufUpsertImpl struct {
	opts    BufferedWriteOptions
	db      Database
	cancel  context.CancelFunc
	upserts chan upsertOp
	flusher chan chan error
	closer  chan chan error
	err     chan error
}

type upsertOp struct {
	query  interface{}
	record interface{}
}

func (bu *anserBufUpsertImpl) start(ctx context.Context) {
	timer := time.NewTimer(bu.opts.Duration)
	defer timer.Stop()
	ops := 0
	bulk := bu.db.C(bu.opts.Collection).Bulk()
	catcher := grip.NewBasicCatcher()

bufferLoop:
	for {
		select {
		case <-ctx.Done():
			if ops > 0 {
				catcher.Add(errors.Errorf("buffered upsert has %d pending operations", ops))
			}

			bu.err <- catcher.Resolve()
			break bufferLoop
		case <-timer.C:
			if ops > 0 {
				_, err := bulk.Run()
				catcher.Add(err)
				ops = 0
			}
			timer.Reset(bu.opts.Duration)
		case op := <-bu.upserts:
			bulk.Upsert(op.query, op.record)
			ops++

			if ops >= bu.opts.Count {
				_, err := bulk.Run()
				catcher.Add(err)
				ops = 0

				if !timer.Stop() {
					<-timer.C
				}

				timer.Reset(bu.opts.Duration)
			}
		case f := <-bu.flusher:
			select {
			case last := <-bu.upserts:
				bulk.Upsert(last.query, last.record)
				ops++
			default:
			}

			if ops > 0 {
				_, err := bulk.Run()
				catcher.Add(err)
				ops = 0

				f <- err
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(bu.opts.Duration)
			}

			close(f)
		case c := <-bu.closer:
			close(bu.upserts)
			for last := range bu.upserts {
				bulk.Upsert(last.query, last.record)
			}

			if ops > 0 {
				_, err := bulk.Run()
				catcher.Add(err)
			}

			c <- catcher.Resolve()
			close(c)
			bu.cancel = nil
			break bufferLoop
		}
	}

	close(bu.err)
}

func (bu *anserBufUpsertImpl) Close() error {
	if bu.cancel == nil {
		return nil
	}

	res := make(chan error)
	bu.closer <- res

	bu.cancel()
	bu.cancel = nil

	return <-res
}

func (bu *anserBufUpsertImpl) Append(doc interface{}) error {
	if doc == nil {
		return errors.New("cannot insert a nil document")
	}

	id, ok := getDocID(doc)
	if !ok {
		return errors.New("could not find document ID")
	}

	bu.upserts <- upsertOp{
		query:  Document{"_id": id},
		record: doc,
	}

	return nil
}

func getDocID(doc interface{}) (interface{}, bool) {
	switch d := doc.(type) {
	case bson.RawD:
		for _, raw := range d {
			if raw.Name == "_id" {
				return raw.Value, true
			}
		}
	case map[string]interface{}:
		id, ok := d["_id"]
		return id, ok
	case Document:
		id, ok := d["_id"]
		return id, ok
	case bson.M:
		id, ok := d["_id"]
		return id, ok
	case map[string]string:
		id, ok := d["_id"]
		return id, ok
	}

	return nil, false

}

func (bu *anserBufUpsertImpl) Flush() error {
	res := make(chan error)
	bu.flusher <- res
	return <-res
}
