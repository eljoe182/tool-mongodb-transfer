package config

import (
	"fmt"
	"net/url"
	"strings"
)

// DatabaseFromURI returns the database name in a MongoDB URI path, or empty if none is set.
func DatabaseFromURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("parse mongodb uri: %w", err)
	}
	if u.Scheme != "mongodb" && u.Scheme != "mongodb+srv" {
		return "", fmt.Errorf("parse mongodb uri: unsupported scheme %q", u.Scheme)
	}

	db := strings.TrimPrefix(u.EscapedPath(), "/")
	if db == "" {
		return "", nil
	}
	if i := strings.IndexByte(db, '/'); i >= 0 {
		db = db[:i]
	}

	decoded, err := url.PathUnescape(db)
	if err != nil {
		return "", fmt.Errorf("decode database name: %w", err)
	}
	return decoded, nil
}

// ResolveDatabases picks origin and destination database names.
// --database overrides both sides; otherwise each URI path is used.
func ResolveDatabases(override, originURI, destURI string) (originDB, destDB string, err error) {
	override = strings.TrimSpace(override)
	if override != "" {
		return override, override, nil
	}

	originDB, err = DatabaseFromURI(originURI)
	if err != nil {
		return "", "", fmt.Errorf("origin uri: %w", err)
	}
	destDB, err = DatabaseFromURI(destURI)
	if err != nil {
		return "", "", fmt.Errorf("destination uri: %w", err)
	}

	var missing []string
	if originDB == "" {
		missing = append(missing, "origin (set --database or include a database in MONGODB_URI_ORIGIN)")
	}
	if destDB == "" {
		missing = append(missing, "destination (set --database or include a database in MONGODB_URI_DESTINATION)")
	}
	if len(missing) > 0 {
		return "", "", fmt.Errorf("missing database for %s", strings.Join(missing, "; "))
	}

	return originDB, destDB, nil
}
