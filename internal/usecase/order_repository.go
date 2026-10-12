package usecase

import (
	"context"
	"orders-clean-architecture/internal/domain"
)

type OrderRepository interface {
	List(ctx context.Context) ([]domain.Order, error)
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
}
