package storage

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/luma434/audittrail/internal/chain"
)

func testDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("AUDITTRAIL_TEST_DSN")
	if dsn == "" {
		t.Skip("AUDITTRAIL_TEST_DSN not set, skipping postgres integration test")
	}
	return dsn
}

func TestPostgres_AppendAndVerifyViaService(t *testing.T) {
	ctx := context.Background()
	pg, err := NewPostgres(ctx, testDSN(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pg.Close()

	svc := chain.NewService(pg)
	now := time.Now()

	e1, err := svc.AppendEvent(ctx, json.RawMessage(`{"event":"genesis"}`), now)
	if err != nil {
		t.Fatalf("append genesis: %v", err)
	}
	e2, err := svc.AppendEvent(ctx, json.RawMessage(`{"event":"second"}`), now.Add(time.Second))
	if err != nil {
		t.Fatalf("append second: %v", err)
	}
	if e2.PrevHash != e1.Hash {
		t.Fatalf("expected e2.PrevHash %q to equal e1.Hash %q", e2.PrevHash, e1.Hash)
	}

	if brokenAt, err := svc.VerifyChain(ctx); err != nil {
		t.Fatalf("expected valid chain, got error at %d: %v", brokenAt, err)
	}

	entries, err := svc.ListEvents(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(entries))
	}
}
