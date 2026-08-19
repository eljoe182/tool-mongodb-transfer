# Mongo Transfer

CLI tool to copy a MongoDB collection from an origin deployment to a destination deployment.

The same collection name is used on both sides. The database can come from `--database` or from each MongoDB URI path. Origin and destination collections must already exist; the tool does not create them.

## Requirements

- Go 1.22+ (developed with Go 1.26)

## Setup

```bash
cp .env.example .env
# Fill in origin and destination MongoDB URIs in .env

go mod download
```

## Environment variables

| Variable | Required | Description |
|---|---|---|
| `MONGODB_URI_ORIGIN` | Yes | MongoDB connection URI for the origin |
| `MONGODB_URI_DESTINATION` | Yes | MongoDB connection URI for the destination |

URI examples:

```env
MONGODB_URI_ORIGIN=mongodb://localhost:27017/app
MONGODB_URI_DESTINATION=mongodb://user:pass@localhost:27018/app?authSource=admin
```

## Usage

```bash
# Database taken from each URI path
go run ./cmd/mongo-transfer --collection users --operation replace

# --database overrides the URI path on origin and destination
go run ./cmd/mongo-transfer --database mydb --collection users --operation add
```

Build a binary:

```bash
go build -o bin/mongo-transfer ./cmd/mongo-transfer
./bin/mongo-transfer --collection users --operation add
```

### Flags

| Flag | Required | Description |
|---|---|---|
| `--database` | No | Database name on origin and destination. If omitted, each URI path is used (`mongodb://host:27017/dbname`) |
| `--collection` | Yes | Collection name on origin and destination |
| `--operation` | Yes | `replace` or `add` |

### Operations

| Operation | Destination missing | Destination exists |
|---|---|---|
| `replace` | Abort | Delete all destination documents (indexes stay), then insert origin documents |
| `add` | Abort | Insert origin documents without deleting existing ones |

Document `_id` values are preserved. Progress is printed to stderr as `done/total (percent%)` after each batch.

If a document fails to insert (for example a duplicate `_id` during `add`), the error is logged with the document `_id` and the rest of the transfer continues.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Every origin document was written |
| `1` | Invalid flags, missing env vars, connection/validation failure, or one or more documents failed |

## How it works

1. Loads URIs from `.env` / the process environment.
2. Resolves databases: `--database` if set, otherwise the path in each URI.
3. Connects to origin and destination and pings both.
4. Verifies the collection exists on origin and on destination.
5. Counts origin documents for the progress total.
6. If `--operation replace`, empties the destination with `DeleteMany`.
7. Reads origin documents in batches of 500 and inserts them unordered on destination.

The origin collection is never modified.

## Tests

```bash
go test ./...
```
