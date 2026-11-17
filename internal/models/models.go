package models

import "time"

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"` // количество на складе
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Order struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
}

type OrderItem struct {
	ID        int `json:"id"`
	OrderID   int `json:"order_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Transaction struct {
	ID        int       `json:"id"`
	OrderID   int       `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CheckoutRequest struct {
	UserEmail string      `json:"user_email"`
	Items     []OrderItem `json:"items"`
}

type OrderDetail struct {
	OrderID           int               `json:"order_id"`
	UserEmail         string            `json:"user_email"`
	TotalPrice        float64           `json:"total_price"`
	CreatedAt         time.Time         `json:"created_at"`
	TransactionStatus string            `json:"transaction_status"`
	TransactionAmount float64           `json:"transaction_amount"`
	Items             []OrderItemDetail `json:"items"`
}

type OrderItemDetail struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

type PopularProduct struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	TotalSold   int64  `json:"total_sold"`
}
