package config

import (
	"strings"
	"testing"
)

func TestDatabaseFromURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr bool
	}{
		{
			name: "database in path",
			uri:  "mongodb://localhost:27017/app",
			want: "app",
		},
		{
			name: "database with query",
			uri:  "mongodb://user:pass@localhost:27017/billing?authSource=admin",
			want: "billing",
		},
		{
			name: "srv uri",
			uri:  "mongodb+srv://cluster.mongodb.net/prod",
			want: "prod",
		},
		{
			name: "no database",
			uri:  "mongodb://localhost:27017",
			want: "",
		},
		{
			name: "trailing slash",
			uri:  "mongodb://localhost:27017/",
			want: "",
		},
		{
			name: "percent encoded name",
			uri:  "mongodb://localhost:27017/my%2Ddb",
			want: "my-db",
		},
		{
			name:    "missing scheme",
			uri:     "localhost:27017/app",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DatabaseFromURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveDatabases(t *testing.T) {
	tests := []struct {
		name      string
		override  string
		originURI string
		destURI   string
		wantOrig  string
		wantDest  string
		wantErr   string
	}{
		{
			name:      "flag overrides both uris",
			override:  "shared",
			originURI: "mongodb://origin:27017/origin-db",
			destURI:   "mongodb://dest:27017/dest-db",
			wantOrig:  "shared",
			wantDest:  "shared",
		},
		{
			name:      "each uri supplies its own database",
			originURI: "mongodb://origin:27017/origin-db",
			destURI:   "mongodb://dest:27017/dest-db",
			wantOrig:  "origin-db",
			wantDest:  "dest-db",
		},
		{
			name:      "origin uri missing database",
			originURI: "mongodb://origin:27017",
			destURI:   "mongodb://dest:27017/dest-db",
			wantErr:   "origin",
		},
		{
			name:      "destination uri missing database",
			originURI: "mongodb://origin:27017/origin-db",
			destURI:   "mongodb://dest:27017",
			wantErr:   "destination",
		},
		{
			name:      "flag fills both when uris have no database",
			override:  "app",
			originURI: "mongodb://origin:27017",
			destURI:   "mongodb://dest:27017",
			wantOrig:  "app",
			wantDest:  "app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOrig, gotDest, err := ResolveDatabases(tt.override, tt.originURI, tt.destURI)
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
			if gotOrig != tt.wantOrig || gotDest != tt.wantDest {
				t.Fatalf("got %q/%q, want %q/%q", gotOrig, gotDest, tt.wantOrig, tt.wantDest)
			}
		})
	}
}
