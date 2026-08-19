package mongo

import (
	"context"
	"testing"
)

func TestConnectInvalidURI(t *testing.T) {
	_, err := Connect(context.Background(), "://not-a-uri")
	if err == nil {
		t.Fatal("expected error for invalid URI, got nil")
	}
}
