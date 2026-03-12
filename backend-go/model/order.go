package model

import (
	"time"
)

type Order struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	TotalAmount   float64   `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status        string    `gorm:"size:20;default:'pending'" json:"status"`
	PaymentMethod string    `gorm:"size:50" json:"payment_method"`
	ShippingAddr  string    `gorm:"type:text" json:"shipping_address"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Items         []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	OrderID   uint    `gorm:"not null;index" json:"order_id"`
	ProductID uint    `gorm:"not null" json:"product_id"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
