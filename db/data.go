package db

import "time"

type ChangeInfo struct {
	Updated    int         // Number of existing documents updated
	Removed    int         // Number of documents removed
	UpsertedId interface{} // Upserted _id field, when not explicitly provided
}

type Index struct {
	Key        []string // Index key fields; prefix name with dash (-) for descending order
	Unique     bool     // Prevent two documents from having the same index key
	DropDups   bool     // Drop documents with the same index key as a previously indexed one
	Background bool     // Build index in background and return immediately
	Sparse     bool     // Only index documents containing the Key fields

	ExpireAfter time.Duration // Periodically delete docs with indexed time.Time older than that.

	Name string // Index name, computed by EnsureIndex

	Bits, Min, Max int // Properties for spatial indexes
}
