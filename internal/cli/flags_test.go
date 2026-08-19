package cli

import (
	"io"
	"strings"
	"testing"

	"mongodb-transfer/internal/transfer"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Flags
		wantErr string
	}{
		{
			name: "replace operation",
			args: []string{"--database", "app", "--collection", "users", "--operation", "replace"},
			want: Flags{Database: "app", Collection: "users", Operation: transfer.OperationReplace},
		},
		{
			name: "add operation is case insensitive",
			args: []string{"--database", "app", "--collection", "users", "--operation", "ADD"},
			want: Flags{Database: "app", Collection: "users", Operation: transfer.OperationAdd},
		},
		{
			name: "database is optional",
			args: []string{"--collection", "users", "--operation", "add"},
			want: Flags{Collection: "users", Operation: transfer.OperationAdd},
		},
		{
			name: "blank database is treated as omitted",
			args: []string{"--database", "  ", "--collection", "users", "--operation", "add"},
			want: Flags{Collection: "users", Operation: transfer.OperationAdd},
		},
		{
			name:    "missing collection",
			args:    []string{"--database", "app", "--operation", "add"},
			wantErr: "--collection",
		},
		{
			name:    "missing operation",
			args:    []string{"--database", "app", "--collection", "users"},
			wantErr: "--operation",
		},
		{
			name:    "invalid operation",
			args:    []string{"--database", "app", "--collection", "users", "--operation", "upsert"},
			wantErr: "invalid --operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlags(tt.args, io.Discard)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseOperation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    transfer.Operation
		wantErr bool
	}{
		{name: "replace", input: "replace", want: transfer.OperationReplace},
		{name: "add", input: "add", want: transfer.OperationAdd},
		{name: "padded", input: " add ", want: transfer.OperationAdd},
		{name: "unknown", input: "merge", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transfer.ParseOperation(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
