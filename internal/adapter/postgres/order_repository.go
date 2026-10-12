package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"orders-clean-architecture/internal/domain"
	"orders-clean-architecture/internal/usecase"
)

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) List(ctx context.Context) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, customer, description, amount, created_at
		FROM orders ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		var order domain.Order
		if err := rows.Scan(&order.ID, &order.Customer, &order.Description, &order.Amount, &order.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (r *PostgresOrderRepository) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO orders (customer, description, amount)
		VALUES ($1, $2, $3)
		RETURNING id, customer, description, amount, created_at
	`, order.Customer, order.Description, order.Amount).Scan(
		&order.ID, &order.Customer, &order.Description, &order.Amount, &order.CreatedAt,
	)
	return order, err
}

var _ usecase.OrderRepository = (*PostgresOrderRepository)(nil)
