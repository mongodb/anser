package db

import (
	"context"
	"strings"
	"time"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type sessionWrapper struct {
	ctx     context.Context
	client  *mongo.Client
	catcher grip.Catcher
}

func (s *sessionWrapper) Clone() Session                   { return s }
func (s *sessionWrapper) Copy() Session                    { return s }
func (s *sessionWrapper) Close()                           { s.catcher.Add(s.client.Disconnect(s.ctx)) }
func (s *sessionWrapper) Error() error                     { return s.catcher.Resolve() }
func (s *sessionWrapper) SetSocketTimeout(d time.Duration) {}
func (s *sessionWrapper) DB(name string) Database {
	return &databaseWrapper{
		ctx:      s.ctx,
		database: s.client.Database(name),
	}
}

type databaseWrapper struct {
	ctx      context.Context
	database *mongo.Database
}

func (d *databaseWrapper) Name() string        { return d.database.Name() }
func (d *databaseWrapper) DropDatabase() error { return errors.WithStack(d.database.Drop(d.ctx)) }
func (d *databaseWrapper) C(coll string) Collection {
	return &collectionWrapper{
		ctx:  d.ctx,
		coll: d.database.Collection(coll),
	}
}

type collectionWrapper struct {
	ctx  context.Context
	coll *mongo.Collection
}

func (c *collectionWrapper) DropCollection() error { return errors.WithStack(c.coll.Drop(c.ctx)) }

func (c *collectionWrapper) Pipe(p interface{}) Results {
	cursor, err := c.coll.Aggregate(c.ctx, p)

	return &resultsWrapper{
		err:    err,
		cursor: cursor,
		ctx:    c.ctx,
	}
}

func (c *collectionWrapper) Find(q interface{}) Query   { return nil }
func (c *collectionWrapper) FindId(q interface{}) Query { return nil }

func (c *collectionWrapper) Count() (int, error)                         { return 0, nil }
func (c *collectionWrapper) Insert(d ...interface{}) error               { return nil }
func (c *collectionWrapper) Update(q interface{}, u interface{}) error   { return nil }
func (c *collectionWrapper) UpdateId(q interface{}, u interface{}) error { return nil }
func (c *collectionWrapper) UpdateAll(q interface{}, u interface{}) (*ChangeInfo, error) {
	return nil, nil
}
func (c *collectionWrapper) Remove(q interface{}) error                               { return nil }
func (c *collectionWrapper) RemoveId(q interface{}) error                             { return nil }
func (c *collectionWrapper) RemoveAll(q interface{}) (*ChangeInfo, error)             { return nil, nil }
func (c *collectionWrapper) Bulk() Bulk                                               { return nil }
func (c *collectionWrapper) Upsert(q interface{}, u interface{}) (*ChangeInfo, error) { return nil, nil }
func (c *collectionWrapper) UpsertId(q interface{}, u interface{}) (*ChangeInfo, error) {
	return nil, nil
}

type resultsWrapper struct {
	ctx    context.Context
	cursor *mongo.Cursor
	err    error
}

func (r *resultsWrapper) All(interface{}) error { return nil }
func (r *resultsWrapper) One(interface{}) error { return nil }

func (r *resultsWrapper) Iter() Iterator {
	catcher := grip.NewCatcher()
	catcher.Add(r.err)
	return &iteratorWrapper{
		ctx:     r.ctx,
		cursor:  r.cursor,
		catcher: catcher,
	}
}

type iteratorWrapper struct {
	ctx        context.Context
	cursor     *mongo.Cursor
	catcher    grip.Catcher
	errChecked bool
}

func (iter *iteratorWrapper) Close() error { return errors.WithStack(iter.cursor.Close(iter.ctx)) }

func (iter *iteratorWrapper) Err() error {
	if !iter.errChecked {
		iter.catcher.Add(iter.cursor.Err())
		iter.errChecked = true
	}

	return iter.catcher.Resolve()
}

func (iter *iteratorWrapper) Next(val interface{}) bool {
	if !iter.cursor.Next(iter.ctx) {
		return false
	}

	iter.catcher.Add(iter.cursor.Decode(val))
	return true
}

type queryWrapper struct {
	ctx        context.Context
	coll       *mongo.Collection
	cursor     *mongo.Cursor
	filter     interface{}
	projection interface{}
	limit      int
	skip       int
	sort       []string
}

func (q *queryWrapper) Limit(l int) Query             { q.limit = l; return q }
func (q *queryWrapper) Select(proj interface{}) Query { q.projection = proj; return q }
func (q *queryWrapper) Sort(keys ...string) Query     { q.sort = append(q.sort, keys...); return q }
func (q *queryWrapper) Skip(s int) Query              { q.skip = s; return q }
func (q *queryWrapper) Count() (int, error) {
	v, err := q.coll.CountDocuments(q.ctx, q.filter)
	return int(v), errors.WithStack(err)
}

func (q *queryWrapper) All(interface{}) error { return nil }
func (q *queryWrapper) One(interface{}) error { return nil }

func (q *queryWrapper) Iter() Iterator {
	if q.cursor != nil {
		return &iteratorWrapper{
			ctx:     q.ctx,
			cursor:  q.cursor,
			catcher: grip.NewCatcher(),
		}
	}

	catcher := grip.NewCatcher()
	opts := options.Find()
	opts.Projection = q.projection

	if q.limit > 0 {
		opts.SetLimit(int64(q.limit))
	}

	if q.skip > 0 {
		opts.SetSkip(int64(q.skip))
	}

	if q.sort != nil {
		sort := bson.D{}

		for _, k := range q.sort {
			if strings.HasPrefix(k, "-") {
				sort = append(sort, bson.E{k[1:], -11})
			} else {
				sort = append(sort, bson.E{k, 1})
			}
		}

		opts.SetSort(sort)
	}

	var err error
	q.cursor, err = q.coll.Find(q.ctx, q.filter, opts)
	catcher.Add(err)
	return &iteratorWrapper{
		ctx:     q.ctx,
		cursor:  q.cursor,
		catcher: catcher,
	}
}
