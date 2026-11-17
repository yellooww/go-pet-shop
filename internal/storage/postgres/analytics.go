package postgres

import (
	"context"
	"fmt"
	"go-pet-shop/internal/models"
	"go-pet-shop/internal/storage"
)

func (s *Storage) GetUserOrderHistory(ctx context.Context, email string) ([]models.OrderDetail, error) {
	const fn = "storage.postgres.analytics.GetUserOrderHistory"

	// Проверяем, существует ли пользователь
	var userExists bool

	err := s.db.QueryRow(ctx,
		`SELECT EXISTS(
            SELECT 1 FROM users WHERE email = $1
        )`,
		email,
	).Scan(&userExists)

	if err != nil {
		return nil, fmt.Errorf("%s: check user: %w", fn, err)
	}

	if !userExists {
		return nil, fmt.Errorf("%s: %w", fn, storage.ErrNotFound)
	}

	// Получаем историю заказов
	rows, err := s.db.Query(ctx,
		`SELECT o.id, u.email, o.total_price, o.created_at, COALESCE(t.status, ''), COALESCE(t.amount, 0),
        COALESCE(oi.product_id, 0), COALESCE(p.name, ''), COALESCE(oi.quantity, 0)
        FROM users u JOIN orders o ON o.user_id = u.id
        LEFT JOIN transactions t ON t.order_id = o.id
        LEFT JOIN order_items oi ON oi.order_id = o.id
        LEFT JOIN products p ON p.id = oi.product_id
        WHERE u.email = $1
        ORDER BY o.created_at DESC, o.id DESC, oi.id ASC`,
		email,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: query history: %w", fn, err)
	}
	defer rows.Close()

	var history []models.OrderDetail
	orderIndexes := make(map[int]int)

	for rows.Next() {
		var orderID int
		var detail models.OrderDetail
		var item models.OrderItemDetail

		err := rows.Scan(
			&orderID,
			&detail.UserEmail,
			&detail.TotalPrice,
			&detail.CreatedAt,
			&detail.TransactionStatus,
			&detail.TransactionAmount,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
		)

		if err != nil {
			return nil, fmt.Errorf("%s: scan history: %w", fn, err)
		}

		// Если заказ встретился впервые, создаём его
		index, exists := orderIndexes[orderID]

		if !exists {
			detail.OrderID = orderID
			detail.Items = make([]models.OrderItemDetail, 0)

			history = append(history, detail)
			index = len(history) - 1

			orderIndexes[orderID] = index
		}

		// Добавляем позицию, если она существует
		if item.ProductID > 0 {
			history[index].Items = append(
				history[index].Items,
				item,
			)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: iterate history: %w", fn, err)
	}

	return history, nil
}

func (s *Storage) GetPopularProducts(ctx context.Context) ([]models.PopularProduct, error) {
	const fn = "storage.postgres.analytics.GetPopularProducts"

	rows, err := s.db.Query(ctx,
		`SELECT p.id, p.name, SUM(oi.quantity) AS total_sold
        FROM products p
        JOIN order_items oi ON oi.product_id = p.id
        GROUP BY p.id, p.name
        ORDER BY total_sold DESC, p.id ASC`,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: query popular products: %w", fn, err)
	}
	defer rows.Close()

	var products []models.PopularProduct

	for rows.Next() {
		var product models.PopularProduct

		err := rows.Scan(
			&product.ProductID,
			&product.ProductName,
			&product.TotalSold,
		)

		if err != nil {
			return nil, fmt.Errorf("%s: scan popular product: %w", fn, err)
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: iterate popular products: %w", fn, err)
	}

	return products, nil
}
