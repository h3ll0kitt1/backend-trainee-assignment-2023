package storage

import (
	"context"

	"github.com/h3ll0kitt1/avitotest/internal/models"
)

type Storage interface {
	// segment
	CreateSegment(ctx context.Context, slug string, PercentageRND int) error
	DeleteSegment(ctx context.Context, slug string) error

	// user
	GetSegmentsByUserID(ctx context.Context, user int64) ([]models.Segment, error)
	UpdateSegmentsByUserID(ctx context.Context, user int64, deleteList []models.Segment, addList []models.Segment) error

	// history
	GetHistory(ctx context.Context, users []int64, days int) ([]models.History, error)
}
