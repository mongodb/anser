// Package mock contains mocked implementations of the interfaces
// defined in the anser package.
package mock

import "github.com/tychoish/anser/db"

type Session struct {
	DBs    map[string]*Database
	URI    string
	closed bool
}

func NewSession() *Session {
	return &Session{
		DBs: make(map[string]*Database),
	}
}

func (s *Session) Clone() db.Session { return s }
func (s *Session) Copy() db.Session  { return s }
func (s *Session) Close()            { s.closed = true }
func (s *Session) DB(n string) db.Database {
	if _, ok := s.DBs[n]; !ok {
		s.DBs[n] = &Database{
			Collections: make(map[string]*Collection),
		}

	}
	return s.DBs[n]
}

type Database struct {
	Collections map[string]*Collection
	DBName      string
}

func (d *Database) Name() string { return d.DBName }
func (d *Database) C(n string) db.Collection {
	if _, ok := d.Collections[n]; !ok {
		d.Collections[n] = &Collection{}
	}

	return d.Collections[n]
}

type Collection struct {
	Name         string
	InsertedDocs []interface{}
}

func (c *Collection) Pipe(p interface{}) db.Pipeline                     { return &Pipeline{Pipe: p} }
func (c *Collection) Find(q interface{}) db.Query                        { return &Query{Query: q} }
func (c *Collection) FindId(q interface{}) db.Query                      { return &Query{Query: q} }
func (c *Collection) Count() (int, error)                                { return len(c.InsertedDocs), nil }
func (c *Collection) Update(q, u interface{}) error                      { return nil }
func (c *Collection) UpdateAll(q, u interface{}) (*db.ChangeInfo, error) { return &db.ChangeInfo{}, nil }
func (c *Collection) UpdateId(id, u interface{}) error                   { return nil }
func (c *Collection) Remove(q interface{}) error                         { return nil }
func (c *Collection) RemoveAll(q interface{}) (*db.ChangeInfo, error)    { return &db.ChangeInfo{}, nil }
func (c *Collection) RemoveId(id interface{}) error                      { return nil }
func (c *Collection) Insert(docs ...interface{}) error                   { c.InsertedDocs = docs; return nil }
func (c *Collection) Upsert(q, u interface{}) (*db.ChangeInfo, error)    { return &db.ChangeInfo{}, nil }
func (c *Collection) UpsertId(id, u interface{}) (*db.ChangeInfo, error) {
	return &db.ChangeInfo{0, 0, id}, nil
}

type Query struct {
	Query    interface{}
	Project  interface{}
	NumLimit int
	NumSkip  int
}

func (q *Query) Count() (int, error)           { return 0, nil }
func (q *Query) Limit(n int) db.Query          { q.NumLimit = n; return q }
func (q *Query) Select(p interface{}) db.Query { q.Project = p; return q }
func (q *Query) Skip(n int) db.Query           { q.NumSkip = n; return q }
func (q *Query) Iter() db.Iterator             { return &Iterator{Query: q} }
func (q *Query) One(r interface{}) error       { return nil }
func (q *Query) All(r interface{}) error       { return nil }
func (q *Query) Sort(keys ...string) db.Query  { return q }

type Iterator struct {
	Query      *Query
	Pipeline   *Pipeline
	ShouldIter bool
	Error      error
}

func (i *Iterator) Next(out interface{}) bool { return i.ShouldIter }
func (i *Iterator) Close() error              { return i.Error }
func (i *Iterator) Err() error                { return i.Error }

type Pipeline struct {
	Pipe  interface{}
	Error error
}

func (p *Pipeline) Iter() db.Iterator       { return &Iterator{Pipeline: p} }
func (p *Pipeline) All(r interface{}) error { return p.Error }
func (p *Pipeline) One(r interface{}) error { return p.Error }
