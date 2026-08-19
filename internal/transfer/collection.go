package transfer

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Cursor streams origin documents one at a time.
type Cursor interface {
	Next(ctx context.Context) bool
	Decode(val any) error
	Err() error
	Close(ctx context.Context) error
}

// DocFailure is a per-document write error that should be logged without aborting.
type DocFailure struct {
	ID  string
	Err error
}

// InsertResult is the outcome of one unordered insert batch.
type InsertResult struct {
	Inserted int
	Failures []DocFailure
}

// Collection is the origin/destination surface used by the transfer engine.
type Collection interface {
	Exists(ctx context.Context) (bool, error)
	Count(ctx context.Context) (int64, error)
	Find(ctx context.Context) (Cursor, error)
	DeleteMany(ctx context.Context) (int64, error)
	InsertMany(ctx context.Context, docs []bson.Raw) (InsertResult, error)
}

// DocumentID returns a printable _id from a raw BSON document.
func DocumentID(doc bson.Raw) string {
	val, err := doc.LookupErr("_id")
	if err != nil {
		return "<missing _id>"
	}
	return val.String()
}

func collectionLabel(name, fallback string) string {
	if name != "" {
		return name
	}
	return fallback
}

func missingCollectionError(kind, name string) error {
	return fmt.Errorf("%s collection %q does not exist", kind, name)
}
