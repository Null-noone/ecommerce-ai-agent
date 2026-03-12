package model

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Email        string    `gorm:"size:100" json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	ParentID  *uint     `gorm:"index" json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Category) TableName() string {
	return "categories"
}

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null;index" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock       int       `gorm:"default:0;not null" json:"stock"`
	CategoryID  *uint     `gorm:"index" json:"category_id"`
	ImageURL    string    `gorm:"size:500" json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Product) TableName() string {
	return "products"
}

type Order struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	UserID        uint        `gorm:"not null;index" json:"user_id"`
	TotalAmount   float64     `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status        string      `gorm:"size:20;default:'pending'" json:"status"`
	PaymentMethod string      `gorm:"size:50" json:"payment_method"`
	ShippingAddr  string      `gorm:"type:text" json:"shipping_address"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Items         []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"not null;index" json:"order_id"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

type ProductEmbedding struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProductID  uint      `gorm:"uniqueIndex;not null" json:"product_id"`
	VectorID   string    `gorm:"size:100" json:"vector_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ProductEmbedding) TableName() string {
	return "product_embeddings"
}

type ChatSession struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"uniqueIndex;size:100;not null" json:"session_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}
