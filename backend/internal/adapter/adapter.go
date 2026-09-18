package adapter

import (
	"context"

	"organizer/backend/internal/domain"
)

// Adapter defines the contract for scanning and extracting domain Items from an AI client or directory.
type Adapter interface {
	Name() string
	Scan(ctx context.Context) ([]domain.Item, error)
}
