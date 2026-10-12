package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"orders-clean-architecture/internal/domain"
)

type fakeOrderRepository struct {
	orders    []domain.Order
	created   domain.Order
	listErr   error
	createErr error
}

func (f *fakeOrderRepository) List(context.Context) ([]domain.Order, error) {
	return f.orders, f.listErr
}

func (f *fakeOrderRepository) Create(_ context.Context, order domain.Order) (domain.Order, error) {
	if f.createErr != nil {
		return domain.Order{}, f.createErr
	}
	f.created = order
	order.ID = 1
	order.CreatedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return order, nil
}

func TestListOrdersReturnsRepositoryOrders(t *testing.T) {
	want := []domain.Order{{ID: 1, Customer: "Cliente", Description: "Material", Amount: 12.5}}
	repo := &fakeOrderRepository{orders: want}

	got, err := NewListOrdersUseCase(repo).Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != want[0].ID || got[0].Customer != want[0].Customer {
		t.Fatalf("Execute() = %#v, want %#v", got, want)
	}
}

func TestListOrdersReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	_, err := NewListOrdersUseCase(&fakeOrderRepository{listErr: wantErr}).Execute(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
}

func TestCreateOrderTrimsFields(t *testing.T) {
	repo := &fakeOrderRepository{}
	got, err := NewCreateOrder(repo).Execute(context.Background(), domain.Order{
		Customer: "  Cliente  ", Description: "  Pedido  ", Amount: 20,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.Customer != "Cliente" || got.Description != "Pedido" {
		t.Fatalf("fields were not trimmed: %#v", got)
	}
}

func TestCreateOrderRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		order domain.Order
	}{
		{name: "missing customer", order: domain.Order{Description: "Pedido"}},
		{name: "missing description", order: domain.Order{Customer: "Cliente"}},
		{name: "negative amount", order: domain.Order{Customer: "Cliente", Description: "Pedido", Amount: -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewCreateOrder(&fakeOrderRepository{}).Execute(context.Background(), tt.order); err == nil {
				t.Fatal("Execute() expected validation error")
			}
		})
	}
}
