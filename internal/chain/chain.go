package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Entry struct {
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	PrevHash  string          `json:"prev_hash"`
	Hash      string          `json:"hash"`
}

const GenesisPrevHash = ""

var ErrChainBroken = errors.New("chain broken")

func computeHash(prevHash string, payload json.RawMessage, timestamp time.Time) (string, error) {
	canonical, err := canonicalize(payload)
	if err != nil {
		return "", fmt.Errorf("canonicalize payload: %w", err)
	}
	h := sha256.New()
	h.Write([]byte(prevHash))
	h.Write(canonical)
	h.Write([]byte(timestamp.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func canonicalize(payload json.RawMessage) ([]byte, error) {
	var v any
	if err := json.Unmarshal(payload, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// now is truncated to microsecond precision: Postgres timestamptz only
// stores microseconds, so hashing at full nanosecond precision would make
// Verify fail on entries that round-tripped through storage even though
// nothing was tampered with.
func Append(prev *Entry, payload json.RawMessage, now time.Time) (*Entry, error) {
	now = now.Truncate(time.Microsecond)
	prevHash := GenesisPrevHash
	if prev != nil {
		prevHash = prev.Hash
	}
	hash, err := computeHash(prevHash, payload, now)
	if err != nil {
		return nil, err
	}
	return &Entry{Payload: payload, Timestamp: now, PrevHash: prevHash, Hash: hash}, nil
}

func Verify(entries []Entry) (brokenAt int, err error) {
	prevHash := GenesisPrevHash
	for i, e := range entries {
		if e.PrevHash != prevHash {
			return i, ErrChainBroken
		}
		wantHash, herr := computeHash(e.PrevHash, e.Payload, e.Timestamp)
		if herr != nil {
			return i, fmt.Errorf("%w: %v", ErrChainBroken, herr)
		}
		if wantHash != e.Hash {
			return i, ErrChainBroken
		}
		prevHash = e.Hash
	}
	return -1, nil
}
