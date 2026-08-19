package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"mongodb-transfer/internal/transfer"
)

// Flags are the command-line inputs.
type Flags struct {
	Database   string
	Collection string
	Operation  transfer.Operation
}

// ParseFlags parses --database, --collection, and --operation from args.
func ParseFlags(args []string, output io.Writer) (Flags, error) {
	fs := flag.NewFlagSet("mongo-transfer", flag.ContinueOnError)
	if output != nil {
		fs.SetOutput(output)
	}

	var database, collection, operation string
	fs.StringVar(&database, "database", "", "Database name on origin and destination (optional; defaults to each URI path)")
	fs.StringVar(&collection, "collection", "", "Collection name on origin and destination")
	fs.StringVar(&operation, "operation", "", "Operation to apply: replace or add")

	if err := fs.Parse(args); err != nil {
		return Flags{}, err
	}

	database = strings.TrimSpace(database)
	collection = strings.TrimSpace(collection)

	var missing []string
	if collection == "" {
		missing = append(missing, "--collection")
	}
	if strings.TrimSpace(operation) == "" {
		missing = append(missing, "--operation")
	}
	if len(missing) > 0 {
		return Flags{}, fmt.Errorf("missing required flags: %s", strings.Join(missing, ", "))
	}

	op, err := transfer.ParseOperation(operation)
	if err != nil {
		return Flags{}, err
	}

	return Flags{
		Database:   database,
		Collection: collection,
		Operation:  op,
	}, nil
}

func isHelp(err error) bool {
	return errors.Is(err, flag.ErrHelp)
}
