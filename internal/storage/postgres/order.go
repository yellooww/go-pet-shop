package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateOrder(ctx context.Context, o models.Order) (int, error) {
	const fn = "storage.postgres.order.CreateOrder"

	var id int
	err := s.db.QueryRow(ctx,
		`INSERT INTO orders (user_id, total_price) VALUES ($1, 0) RETURNING id`,
		o.UserID).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) AddOrderItem(ctx context.Context, oi models.OrderItem) (int, error) {
	const fn = "storage.postgres.order.AddOrderItem"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	defer tx.Rollback(ctx)

	// Проверяем существование заказа
	var orderID int
	err = tx.QueryRow(ctx,
		`SELECT id FROM orders WHERE id = $1`,
		oi.OrderID,
	).Scan(&orderID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	// Проверяем существование товара
	var productID int
	err = tx.QueryRow(ctx,
		`SELECT id FROM products WHERE id = $1`,
		oi.ProductID,
	).Scan(&productID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	var itemID int
	err = tx.QueryRow(ctx,
		`INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3) RETURNING id`,
		oi.OrderID, oi.ProductID, oi.Quantity).Scan(&itemID)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE orders
         SET total_price = total_price +
             (SELECT price * $1 FROM products WHERE id = $2)
         WHERE id = $3`,
		oi.Quantity, oi.ProductID, oi.OrderID,
	)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return itemID, nil
}

func (s *Storage) GetOrderByID(ctx context.Context, id int) (models.Order, error) {
	const fn = "storage.postgres.order.GetOrderByID"

	var o models.Order
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, total_price, created_at FROM orders WHERE id = $1`,
		id).Scan(&o.ID, &o.UserID, &o.TotalPrice, &o.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return models.Order{}, fmt.Errorf("%s: %w", fn, err)
	}

	return o, nil
}

func (s *Storage) GetOrdersByUserEmail(ctx context.Context, email string) ([]models.Order, error) {
	const fn = "storage.postgres.order.GetOrdersByUserEmail"

	rows, err := s.db.Query(ctx,
		`SELECT o.id, o.user_id, o.total_price, o.created_at FROM orders o JOIN users u ON u.id = o.user_id
		WHERE u.email = $1 ORDER BY o.created_at DESC;`, email)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.TotalPrice, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return orders, nil
}

func (s *Storage) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	const fn = "storage.postgres.order.GetOrderItemsByOrderID"

	rows, err := s.db.Query(ctx,
		`SELECT id, order_id, product_id, quantity FROM order_items WHERE order_id = $1`,
		orderID)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	var orderItems []models.OrderItem
	for rows.Next() {
		var oi models.OrderItem
		if err := rows.Scan(&oi.ID, &oi.OrderID, &oi.ProductID, &oi.Quantity); err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}
		orderItems = append(orderItems, oi)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return orderItems, nil
}

func (s *Storage) PlaceOrder(ctx context.Context, userEmail string, items []models.OrderItem) (int, error) {
	const fn = "storage.postgres.order.PlaceOrder"

	if len(items) == 0 {
		return 0, fmt.Errorf("%s: %w", fn, storage.ErrInvalidInput)
	}

	for _, item := range items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrInvalidInput)
		}
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}
	defer tx.Rollback(ctx)

	var userID int
	err = tx.QueryRow(ctx,
		`SELECT id FROM users WHERE email = $1`,
		userEmail).Scan(&userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	var orderID int

	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, total_price) VALUES ($1, 0) RETURNING id`,
		userID).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	for _, item := range items {
		var productExists bool

		err = tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`,
			item.ProductID).Scan(&productExists)

		if err != nil {
			return 0, fmt.Errorf("%s: %w", fn, err)
		}
		if !productExists {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
		}

		result, err := tx.Exec(ctx,
			`UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1`,
			item.Quantity, item.ProductID)

		if err != nil {
			return 0, fmt.Errorf("%s: %w", fn, err)
		}

		affectedRows := result.RowsAffected()
		if affectedRows == 0 {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrInsufficientStock)
		}

		_, err = tx.Exec(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)`,
			orderID, item.ProductID, item.Quantity)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", fn, err)
		}

		_, err = tx.Exec(ctx,
			`UPDATE orders
			SET total_price = total_price +
				(SELECT price * $1 FROM products WHERE id = $2)
			WHERE id = $3`,
			item.Quantity, item.ProductID, orderID,
		)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", fn, err)
		}
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO transactions (order_id, amount, status) VALUES ($1, 
		(SELECT total_price FROM orders WHERE id = $1), $2)`, orderID, "pending")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return orderID, nil
}
