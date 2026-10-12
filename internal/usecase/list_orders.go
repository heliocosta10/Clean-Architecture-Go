package usecase

import (
	"context"
	"orders-clean-architecture/internal/domain"
)

type ListOrders struct {
	repository OrderRepository
}

func NewListOrders(repository OrderRepository) *ListOrders {
	return &ListOrders{repository: repository}
}

func (uc *ListOrders) Execute(ctx context.Context) ([]domain.Order, error) {
	return uc.repository.List(ctx)
}

type ListOrdersUseCase = ListOrders

func NewListOrdersUseCase(repository OrderRepository) *ListOrdersUseCase {
	return NewListOrders(repository)
}
