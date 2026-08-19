package transfer

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteProgress(t *testing.T) {
	tests := []struct {
		name  string
		done  int64
		total int64
		want  string
	}{
		{name: "midway", done: 45, total: 100, want: "\r45/100 (45%)"},
		{name: "complete", done: 10, total: 10, want: "\r10/10 (100%)"},
		{name: "empty collection", done: 0, total: 0, want: "\r0/0 (100%)"},
		{name: "caps at 100", done: 12, total: 10, want: "\r12/10 (100%)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			WriteProgress(&buf, tt.done, tt.total)
			if got := buf.String(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogDocumentError(t *testing.T) {
	var buf bytes.Buffer
	LogDocumentError(&buf, `"abc"`, errString("duplicate key"))
	got := buf.String()
	if !strings.Contains(got, `"abc"`) || !strings.Contains(got, "duplicate key") {
		t.Fatalf("unexpected log line: %q", got)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
