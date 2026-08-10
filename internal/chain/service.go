package chain

import (
	"context"
	"encoding/json"
	"time"
)

// Service orchestrates appending to and verifying a chain backed by a
// Repository. It is the only part of this package that touches storage,
// and it does so exclusively through the Repository interface.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// AppendEvent loads the current chain head from the repository, computes
// the next entry, and persists it.
func (s *Service) AppendEvent(ctx context.Context, payload json.RawMessage, now time.Time) (*Entry, error) {
	prev, err := s.repo.Latest(ctx)
	if err != nil {
		return nil, err
	}
	entry, err := Append(prev, payload, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Insert(ctx, *entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// ListEvents returns a paginated page of entries as stored.
func (s *Service) ListEvents(ctx context.Context, limit, offset int) ([]Entry, error) {
	return s.repo.List(ctx, limit, offset)
}

// VerifyChain loads every stored entry and walks the whole chain,
// failing closed at the first broken link.
func (s *Service) VerifyChain(ctx context.Context) (brokenAt int, err error) {
	entries, err := s.repo.All(ctx)
	if err != nil {
		return -1, err
	}
	return Verify(entries)
}
