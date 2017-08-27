package db

type Session interface {
	Clone() Session
	Copy() Session
	Close()
	DB(string) Database
}

type Database interface {
	Name() string
	C(string) Collection
}

type Collection interface {
	// Indexes() ([]Index, error)

	Pipe(interface{}) Pipeline
	Find(interface{}) Query
	FindId(interface{}) Query
	Count() (int, error)
	Insert(...interface{}) error
	Upsert(interface{}, interface{}) (*ChangeInfo, error)
	UpsertId(interface{}, interface{}) (*ChangeInfo, error)
	Update(interface{}, interface{}) error
	UpdateId(interface{}, interface{}) error
	UpdateAll(interface{}, interface{}) (*ChangeInfo, error)
	Remove(interface{}) error
	RemoveId(interface{}) error
	RemoveAll(interface{}) (*ChangeInfo, error)
}

type Query interface {
	Count() (int, error)
	Limit(int) Query
	Select(interface{}) Query
	Skip(n int) Query
	Iter() Iterator
	One(interface{}) error
	All(interface{}) error
	Sort(...string) Query
}

// Iterator is a more narrow subset of mgo's Iter type that
// provides the opportunity to mock results, and avoids a strict
// dependency between mgo and migrations definitions.
type Iterator interface {
	Next(interface{}) bool
	Close() error
	Err() error
}

type Pipeline interface {
	All(interface{}) error
	One(interface{}) error
	Iter() Iterator
}
