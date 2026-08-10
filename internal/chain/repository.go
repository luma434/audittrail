package chain

import "context"

// Repository persists and retrieves chain entries. Implementations must
// enforce insert-only semantics (no UPDATE/DELETE) so the hash chain
// stays tamper-evident; chain itself never issues those operations.
type Repository interface {
	// Latest returns the most recently inserted entry, or nil if the
	// chain is empty.
	Latest(ctx context.Context) (*Entry, error)
	// Insert appends a single entry.
	Insert(ctx context.Context, e Entry) error
	// List returns entries in chain order, paginated.
	List(ctx context.Context, limit, offset int) ([]Entry, error)
	// All returns every entry in chain order, for full-chain verification.
	All(ctx context.Context) ([]Entry, error)
}
