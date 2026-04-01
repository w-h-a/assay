package verdictstore

import "context"

// VerdictStore is the port interface for storing and retrieving verdicts.
type VerdictStore interface {
	Store(ctx context.Context, verdict *Verdict) error
	Get(ctx context.Context, specHash string) (*Verdict, error)
	List(ctx context.Context) ([]*Verdict, error)
}
