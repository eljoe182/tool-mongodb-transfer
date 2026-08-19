package cli

import (
	"context"
	"fmt"
	"io"

	"mongodb-transfer/internal/config"
	"mongodb-transfer/internal/mongo"
	"mongodb-transfer/internal/transfer"
)

// Run parses flags, connects to origin and destination, and transfers the collection.
func Run(ctx context.Context, args []string, stderr, stdout io.Writer) int {
	flags, err := ParseFlags(args, stderr)
	if err != nil {
		if isHelp(err) {
			return 0
		}
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	originDB, destDB, err := config.ResolveDatabases(flags.Database, cfg.OriginURI, cfg.DestURI)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	originClient, err := mongo.Connect(ctx, cfg.OriginURI)
	if err != nil {
		fmt.Fprintf(stderr, "Error: origin connection: %v\n", err)
		return 1
	}
	defer func() {
		_ = originClient.Disconnect(ctx)
	}()

	destClient, err := mongo.Connect(ctx, cfg.DestURI)
	if err != nil {
		fmt.Fprintf(stderr, "Error: destination connection: %v\n", err)
		return 1
	}
	defer func() {
		_ = destClient.Disconnect(ctx)
	}()

	origin := originClient.Collection(originDB, flags.Collection)
	dest := destClient.Collection(destDB, flags.Collection)

	result, err := transfer.Run(ctx, transfer.Params{
		Origin:     origin,
		Dest:       dest,
		OriginName: flags.Collection,
		DestName:   flags.Collection,
		Operation:  flags.Operation,
		Progress: func(done, total int64) {
			transfer.WriteProgress(stderr, done, total)
		},
		OnError: func(id string, err error) {
			fmt.Fprint(stderr, "\n")
			transfer.LogDocumentError(stderr, id, err)
		},
	})
	fmt.Fprint(stderr, "\n")

	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		if result.Transferred > 0 || result.Failed > 0 {
			fmt.Fprintf(stdout, "Done. transferred=%d failed=%d\n", result.Transferred, result.Failed)
		}
		return 1
	}

	fmt.Fprintf(stdout, "Done. transferred=%d failed=%d\n", result.Transferred, result.Failed)
	return 0
}
