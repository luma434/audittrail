package chain

import (
	"encoding/json"
	"testing"
	"time"
)

func mustPayload(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return b
}

func TestAppendAndVerify_HappyPath(t *testing.T) {
	now := time.Now()
	e1, err := Append(nil, mustPayload(t, map[string]string{"event": "genesis"}), now)
	if err != nil {
		t.Fatalf("append genesis: %v", err)
	}
	e2, err := Append(e1, mustPayload(t, map[string]string{"event": "second"}), now.Add(time.Second))
	if err != nil {
		t.Fatalf("append second: %v", err)
	}

	if brokenAt, err := Verify([]Entry{*e1, *e2}); err != nil {
		t.Fatalf("expected valid chain, got error at %d: %v", brokenAt, err)
	}
}

func TestVerify_EmptyChain(t *testing.T) {
	if brokenAt, err := Verify(nil); err != nil {
		t.Fatalf("expected empty chain to be valid, got error at %d: %v", brokenAt, err)
	}
}

func TestVerify_GenesisEntry(t *testing.T) {
	e1, err := Append(nil, mustPayload(t, map[string]string{"event": "genesis"}), time.Now())
	if err != nil {
		t.Fatalf("append genesis: %v", err)
	}
	if brokenAt, err := Verify([]Entry{*e1}); err != nil {
		t.Fatalf("expected valid single-entry chain, got error at %d: %v", brokenAt, err)
	}
}

func TestVerify_DetectsTamperedPayload(t *testing.T) {
	now := time.Now()
	e1, _ := Append(nil, mustPayload(t, map[string]string{"event": "genesis"}), now)
	e2, _ := Append(e1, mustPayload(t, map[string]string{"event": "second"}), now.Add(time.Second))

	tampered := *e2
	tampered.Payload = mustPayload(t, map[string]string{"event": "tampered"})

	brokenAt, err := Verify([]Entry{*e1, tampered})
	if err == nil {
		t.Fatal("expected tampering to be detected")
	}
	if brokenAt != 1 {
		t.Fatalf("expected break at index 1, got %d", brokenAt)
	}
}

func TestVerify_DetectsTamperedHash(t *testing.T) {
	now := time.Now()
	e1, _ := Append(nil, mustPayload(t, map[string]string{"event": "genesis"}), now)
	e2, _ := Append(e1, mustPayload(t, map[string]string{"event": "second"}), now.Add(time.Second))

	tampered := *e2
	tampered.Hash = "0000000000000000000000000000000000000000000000000000000000000000"

	brokenAt, err := Verify([]Entry{*e1, tampered})
	if err == nil {
		t.Fatal("expected hash tampering to be detected")
	}
	if brokenAt != 1 {
		t.Fatalf("expected break at index 1, got %d", brokenAt)
	}
}

func TestVerify_DetectsBrokenPrevHash(t *testing.T) {
	now := time.Now()
	e1, _ := Append(nil, mustPayload(t, map[string]string{"event": "genesis"}), now)
	e2, _ := Append(e1, mustPayload(t, map[string]string{"event": "second"}), now.Add(time.Second))
	e3, _ := Append(e2, mustPayload(t, map[string]string{"event": "third"}), now.Add(2*time.Second))

	brokenAt, err := Verify([]Entry{*e1, *e3})
	if err == nil {
		t.Fatal("expected deleted entry (broken prev_hash link) to be detected")
	}
	if brokenAt != 1 {
		t.Fatalf("expected break at index 1, got %d", brokenAt)
	}
}

func TestAppend_InvalidPayload(t *testing.T) {
	_, err := Append(nil, json.RawMessage(`{not valid json`), time.Now())
	if err == nil {
		t.Fatal("expected invalid JSON payload to be rejected")
	}
}
