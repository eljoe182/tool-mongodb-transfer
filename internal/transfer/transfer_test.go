package transfer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type fakeCursor struct {
	docs []bson.Raw
	idx  int
	err  error
}

func (c *fakeCursor) Next(context.Context) bool {
	if c.idx >= len(c.docs) {
		return false
	}
	c.idx++
	return true
}

func (c *fakeCursor) Decode(val any) error {
	ptr, ok := val.(*bson.Raw)
	if !ok {
		return fmt.Errorf("unexpected decode type %T", val)
	}
	*ptr = c.docs[c.idx-1]
	return nil
}

func (c *fakeCursor) Err() error { return c.err }

func (c *fakeCursor) Close(context.Context) error { return nil }

type fakeCollection struct {
	exists       bool
	existsErr    error
	countErr     error
	deleteErr    error
	insertErr    error
	docs         []bson.Raw
	deleted      bool
	inserts      []bson.Raw
	failIDs      map[string]error
	insertCalled bool
}

func (f *fakeCollection) Exists(context.Context) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeCollection) Count(context.Context) (int64, error) {
	if f.countErr != nil {
		return 0, f.countErr
	}
	return int64(len(f.docs)), nil
}

func (f *fakeCollection) Find(context.Context) (Cursor, error) {
	return &fakeCursor{docs: f.docs}, nil
}

func (f *fakeCollection) DeleteMany(context.Context) (int64, error) {
	f.deleted = true
	if f.deleteErr != nil {
		return 0, f.deleteErr
	}
	return int64(len(f.inserts)), nil
}

func (f *fakeCollection) InsertMany(_ context.Context, docs []bson.Raw) (InsertResult, error) {
	f.insertCalled = true
	if f.insertErr != nil {
		return InsertResult{}, f.insertErr
	}
	var result InsertResult
	for _, doc := range docs {
		id := DocumentID(doc)
		if err, ok := f.failIDs[id]; ok {
			result.Failures = append(result.Failures, DocFailure{ID: id, Err: err})
			continue
		}
		f.inserts = append(f.inserts, doc)
		result.Inserted++
	}
	return result, nil
}

func mustRaw(t *testing.T, id string) bson.Raw {
	t.Helper()
	raw, err := bson.Marshal(bson.D{{Key: "_id", Value: id}, {Key: "n", Value: id}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestRunMissingCollectionsAbortBeforeWrite(t *testing.T) {
	ctx := context.Background()
	docs := []bson.Raw{mustRaw(t, "a")}

	tests := []struct {
		name   string
		origin *fakeCollection
		dest   *fakeCollection
		want   string
	}{
		{
			name:   "origin missing",
			origin: &fakeCollection{exists: false, docs: docs},
			dest:   &fakeCollection{exists: true},
			want:   `origin collection "users" does not exist`,
		},
		{
			name:   "destination missing",
			origin: &fakeCollection{exists: true, docs: docs},
			dest:   &fakeCollection{exists: false},
			want:   `destination collection "users" does not exist`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Run(ctx, Params{
				Origin:     tt.origin,
				Dest:       tt.dest,
				OriginName: "users",
				DestName:   "users",
				Operation:  OperationReplace,
			})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.want {
				t.Fatalf("got %q, want %q", err.Error(), tt.want)
			}
			if tt.dest.deleted || tt.dest.insertCalled {
				t.Fatal("destination was written despite missing collection")
			}
		})
	}
}

func TestRunReplaceEmptiesThenInserts(t *testing.T) {
	ctx := context.Background()
	origin := &fakeCollection{
		exists: true,
		docs:   []bson.Raw{mustRaw(t, "a"), mustRaw(t, "b")},
	}
	dest := &fakeCollection{exists: true}

	result, err := Run(ctx, Params{
		Origin:    origin,
		Dest:      dest,
		Operation: OperationReplace,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dest.deleted {
		t.Fatal("replace did not empty destination")
	}
	if result.Transferred != 2 || result.Failed != 0 {
		t.Fatalf("result = %+v, want transferred=2 failed=0", result)
	}
	if len(dest.inserts) != 2 {
		t.Fatalf("inserted %d docs, want 2", len(dest.inserts))
	}
}

func TestRunAddDoesNotEmptyDestination(t *testing.T) {
	ctx := context.Background()
	origin := &fakeCollection{
		exists: true,
		docs:   []bson.Raw{mustRaw(t, "a")},
	}
	dest := &fakeCollection{exists: true}

	_, err := Run(ctx, Params{
		Origin:    origin,
		Dest:      dest,
		Operation: OperationAdd,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest.deleted {
		t.Fatal("add emptied destination")
	}
	if len(dest.inserts) != 1 {
		t.Fatalf("inserted %d docs, want 1", len(dest.inserts))
	}
}

func TestRunAddLogsDuplicateAndContinues(t *testing.T) {
	ctx := context.Background()
	dupID := DocumentID(mustRaw(t, "b"))
	origin := &fakeCollection{
		exists: true,
		docs:   []bson.Raw{mustRaw(t, "a"), mustRaw(t, "b"), mustRaw(t, "c")},
	}
	dest := &fakeCollection{
		exists:  true,
		failIDs: map[string]error{dupID: errors.New("E11000 duplicate key")},
	}

	var logged []string
	result, err := Run(ctx, Params{
		Origin:    origin,
		Dest:      dest,
		Operation: OperationAdd,
		OnError: func(id string, err error) {
			logged = append(logged, id)
		},
	})
	if err == nil {
		t.Fatal("expected error because a document failed")
	}
	if result.Transferred != 2 || result.Failed != 1 {
		t.Fatalf("result = %+v, want transferred=2 failed=1", result)
	}
	if len(logged) != 1 || logged[0] != dupID {
		t.Fatalf("logged = %v, want [%s]", logged, dupID)
	}
	if dest.deleted {
		t.Fatal("add emptied destination")
	}
}

func TestRunReportsProgressPerBatch(t *testing.T) {
	ctx := context.Background()
	origin := &fakeCollection{
		exists: true,
		docs:   []bson.Raw{mustRaw(t, "a"), mustRaw(t, "b"), mustRaw(t, "c")},
	}
	dest := &fakeCollection{exists: true}

	var reports [][2]int64
	_, err := Run(ctx, Params{
		Origin:    origin,
		Dest:      dest,
		Operation: OperationAdd,
		BatchSize: 2,
		Progress: func(done, total int64) {
			reports = append(reports, [2]int64{done, total})
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := [][2]int64{{0, 3}, {2, 3}, {3, 3}}
	if len(reports) != len(want) {
		t.Fatalf("reports = %v, want %v", reports, want)
	}
	for i := range want {
		if reports[i] != want[i] {
			t.Fatalf("report[%d] = %v, want %v", i, reports[i], want[i])
		}
	}
}

func TestRunEmptyOriginReplaceStillDeletes(t *testing.T) {
	ctx := context.Background()
	origin := &fakeCollection{exists: true}
	dest := &fakeCollection{exists: true}

	result, err := Run(ctx, Params{
		Origin:    origin,
		Dest:      dest,
		Operation: OperationReplace,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dest.deleted {
		t.Fatal("replace of empty origin did not empty destination")
	}
	if dest.insertCalled {
		t.Fatal("insert called for empty origin")
	}
	if result.Transferred != 0 {
		t.Fatalf("transferred = %d, want 0", result.Transferred)
	}
}
