package transfer

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const defaultBatchSize = 500

// Params configures a collection transfer.
type Params struct {
	Origin     Collection
	Dest       Collection
	OriginName string
	DestName   string
	Operation  Operation
	BatchSize  int
	Progress   Reporter
	OnError    func(id string, err error)
}

// Result summarizes a completed transfer attempt.
type Result struct {
	Transferred int64
	Failed      int64
}

func (p Params) batchSize() int {
	if p.BatchSize <= 0 {
		return defaultBatchSize
	}
	return p.BatchSize
}

func (p Params) report(done, total int64) {
	if p.Progress != nil {
		p.Progress(done, total)
	}
}

func (p Params) logError(id string, err error) {
	if p.OnError != nil {
		p.OnError(id, err)
	}
}

// Run copies documents from origin to destination according to Params.Operation.
func Run(ctx context.Context, p Params) (Result, error) {
	var result Result

	originExists, err := p.Origin.Exists(ctx)
	if err != nil {
		return result, fmt.Errorf("check origin collection: %w", err)
	}
	if !originExists {
		return result, missingCollectionError("origin", collectionLabel(p.OriginName, "origin"))
	}

	destExists, err := p.Dest.Exists(ctx)
	if err != nil {
		return result, fmt.Errorf("check destination collection: %w", err)
	}
	if !destExists {
		return result, missingCollectionError("destination", collectionLabel(p.DestName, "destination"))
	}

	total, err := p.Origin.Count(ctx)
	if err != nil {
		return result, fmt.Errorf("count origin documents: %w", err)
	}

	if p.Operation == OperationReplace {
		if _, err := p.Dest.DeleteMany(ctx); err != nil {
			return result, fmt.Errorf("empty destination collection: %w", err)
		}
	}

	p.report(0, total)
	if total == 0 {
		return result, nil
	}

	cursor, err := p.Origin.Find(ctx)
	if err != nil {
		return result, fmt.Errorf("find origin documents: %w", err)
	}
	defer func() {
		_ = cursor.Close(context.Background())
	}()

	batchSize := p.batchSize()
	batch := make([]bson.Raw, 0, batchSize)

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		inserted, failed, err := insertBatch(ctx, p, batch)
		result.Transferred += inserted
		result.Failed += failed
		p.report(result.Transferred+result.Failed, total)
		batch = batch[:0]
		return err
	}

	for cursor.Next(ctx) {
		var doc bson.Raw
		if err := cursor.Decode(&doc); err != nil {
			return result, fmt.Errorf("decode origin document: %w", err)
		}
		batch = append(batch, doc)
		if len(batch) >= batchSize {
			if err := flush(); err != nil {
				return result, err
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return result, fmt.Errorf("iterate origin documents: %w", err)
	}
	if err := flush(); err != nil {
		return result, err
	}

	if result.Failed > 0 {
		return result, fmt.Errorf("%d document(s) failed", result.Failed)
	}
	return result, nil
}

func insertBatch(ctx context.Context, p Params, batch []bson.Raw) (int64, int64, error) {
	res, err := p.Dest.InsertMany(ctx, batch)
	if err != nil {
		return 0, 0, fmt.Errorf("insert documents: %w", err)
	}
	for _, failure := range res.Failures {
		p.logError(failure.ID, failure.Err)
	}
	return int64(res.Inserted), int64(len(res.Failures)), nil
}
