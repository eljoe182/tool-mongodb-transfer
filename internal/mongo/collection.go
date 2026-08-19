package mongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"mongodb-transfer/internal/transfer"
)

// Collection wraps a driver collection and implements transfer.Collection.
type Collection struct {
	db   *mongo.Database
	coll *mongo.Collection
}

var _ transfer.Collection = (*Collection)(nil)

// Exists reports whether the collection is present in the database.
func (c *Collection) Exists(ctx context.Context) (bool, error) {
	names, err := c.db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: c.coll.Name()}})
	if err != nil {
		return false, fmt.Errorf("list collections: %w", err)
	}
	return len(names) > 0, nil
}

// Count returns the number of documents in the collection.
func (c *Collection) Count(ctx context.Context) (int64, error) {
	n, err := c.coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}
	return n, nil
}

// Find returns a cursor over every document in the collection.
func (c *Collection) Find(ctx context.Context) (transfer.Cursor, error) {
	cur, err := c.coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}
	return &cursor{inner: cur}, nil
}

// DeleteMany removes every document, leaving indexes in place.
func (c *Collection) DeleteMany(ctx context.Context) (int64, error) {
	res, err := c.coll.DeleteMany(ctx, bson.D{})
	if err != nil {
		return 0, fmt.Errorf("delete many: %w", err)
	}
	return res.DeletedCount, nil
}

// InsertMany writes a batch unordered so a single document failure does not abort the rest.
func (c *Collection) InsertMany(ctx context.Context, docs []bson.Raw) (transfer.InsertResult, error) {
	var result transfer.InsertResult
	if len(docs) == 0 {
		return result, nil
	}

	payload := make([]any, len(docs))
	for i, doc := range docs {
		payload[i] = doc
	}

	res, err := c.coll.InsertMany(ctx, payload, options.InsertMany().SetOrdered(false))
	if res != nil {
		result.Inserted = len(res.InsertedIDs)
	}
	if err == nil {
		return result, nil
	}

	var bwe mongo.BulkWriteException
	if !errors.As(err, &bwe) {
		return result, fmt.Errorf("insert many: %w", err)
	}

	for _, we := range bwe.WriteErrors {
		id := "<unknown>"
		if we.Index >= 0 && we.Index < len(docs) {
			id = transfer.DocumentID(docs[we.Index])
		}
		result.Failures = append(result.Failures, transfer.DocFailure{
			ID:  id,
			Err: we.WriteError,
		})
	}
	if res == nil {
		result.Inserted = len(docs) - len(result.Failures)
	}
	return result, nil
}

type cursor struct {
	inner *mongo.Cursor
}

func (c *cursor) Next(ctx context.Context) bool {
	return c.inner.Next(ctx)
}

func (c *cursor) Decode(val any) error {
	return c.inner.Decode(val)
}

func (c *cursor) Err() error {
	return c.inner.Err()
}

func (c *cursor) Close(ctx context.Context) error {
	return c.inner.Close(ctx)
}
