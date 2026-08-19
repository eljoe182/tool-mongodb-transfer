package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const serverSelectionTimeout = 10 * time.Second

// Client wraps a MongoDB client used by the CLI.
type Client struct {
	inner *mongo.Client
}

// Connect creates a client and pings the deployment so connection errors fail fast.
func Connect(ctx context.Context, uri string) (*Client, error) {
	opts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(serverSelectionTimeout)

	inner, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, serverSelectionTimeout)
	defer cancel()
	if err := inner.Ping(pingCtx, nil); err != nil {
		_ = inner.Disconnect(ctx)
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Client{inner: inner}, nil
}

// Collection returns a handle for database/collection.
func (c *Client) Collection(database, collection string) *Collection {
	db := c.inner.Database(database)
	return &Collection{
		db:   db,
		coll: db.Collection(collection),
	}
}

// Disconnect closes the client. Safe to call on a nil receiver.
func (c *Client) Disconnect(ctx context.Context) error {
	if c == nil || c.inner == nil {
		return nil
	}
	if err := c.inner.Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnect: %w", err)
	}
	return nil
}
