package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestRunHelpExitsZero(t *testing.T) {
	code := Run(context.Background(), []string{"-h"}, io.Discard, io.Discard)
	if code != 0 {
		t.Fatalf("help exit = %d, want 0", code)
	}
}

func TestRunMissingFlagsExitsOne(t *testing.T) {
	var stderr bytes.Buffer
	code := Run(context.Background(), []string{"--database", "app", "--collection", "users"}, &stderr, io.Discard)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "--operation") {
		t.Fatalf("stderr = %q, want --operation", stderr.String())
	}
}
