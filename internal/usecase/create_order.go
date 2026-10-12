package usecase

import (
	"context"
	"errors"
	"strings"

	"orders-clean-architecture/internal/domain"
)

type CreateOrder struct {
	repository OrderRepository
}

func NewCreateOrder(repository OrderRepository) *CreateOrder {
	return &CreateOrder{repository: repository}
}

func (uc *CreateOrder) Execute(ctx context.Context, order domain.Order) (domain.Order, error) {
	order.Customer = strings.TrimSpace(order.Customer)
	order.Description = strings.TrimSpace(order.Description)
	if order.Customer == "" {
		return domain.Order{}, errors.New("customer is required")
	}
	if order.Description == "" {
		return domain.Order{}, errors.New("description is required")
	}
	if order.Amount < 0 {
		return domain.Order{}, errors.New("amount must be zero or greater")
	}
	return uc.repository.Create(ctx, order)
}
