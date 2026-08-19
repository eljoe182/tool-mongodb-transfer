package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr string
	}{
		{
			name: "both uris present",
			env: map[string]string{
				EnvOriginURI: "mongodb://origin:27017",
				EnvDestURI:   "mongodb://dest:27017",
			},
			want: Config{
				OriginURI: "mongodb://origin:27017",
				DestURI:   "mongodb://dest:27017",
			},
		},
		{
			name: "trims whitespace",
			env: map[string]string{
				EnvOriginURI: "  mongodb://origin:27017  ",
				EnvDestURI:   " mongodb://dest:27017",
			},
			want: Config{
				OriginURI: "mongodb://origin:27017",
				DestURI:   "mongodb://dest:27017",
			},
		},
		{
			name:    "missing origin",
			env:     map[string]string{EnvDestURI: "mongodb://dest:27017"},
			wantErr: EnvOriginURI,
		},
		{
			name:    "missing destination",
			env:     map[string]string{EnvOriginURI: "mongodb://origin:27017"},
			wantErr: EnvDestURI,
		},
		{
			name:    "missing both",
			env:     map[string]string{},
			wantErr: EnvOriginURI,
		},
		{
			name: "blank origin is missing",
			env: map[string]string{
				EnvOriginURI: "   ",
				EnvDestURI:   "mongodb://dest:27017",
			},
			wantErr: EnvOriginURI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(func(key string) string {
				return tt.env[key]
			})
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
