package types

// ==================== Auth ====================

type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResp struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ==================== Product ====================

type GetProductsReq struct {
	Page     int `json:"page,default=1"`
	PageSize int `json:"page_size,default=10"`
}

type GetProductReq struct {
	ID uint `json:"id"`
}

type ProductInfo struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	CategoryID  uint    `json:"category_id,omitempty"`
	ImageURL    string  `json:"image_url"`
	CreatedAt   string  `json:"created_at"`
}

// ==================== Search ====================

type SearchReq struct {
	Query string `json:"q"`
	Page  int    `json:"page,default=1"`
	Limit int    `json:"limit,default=10"`
}

type SearchResp struct {
	Products []ProductInfo `json:"products"`
	Query    string        `json:"query"`
	Total    int           `json:"total"`
}

// ==================== Order ====================

type CreateOrderReq struct {
	Items []OrderItemReq `json:"items"`
}

type OrderItemReq struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type OrderInfo struct {
	ID            uint          `json:"id"`
	UserID        uint          `json:"user_id"`
	TotalAmount   float64       `json:"total_amount"`
	Status        string        `json:"status"`
	PaymentMethod string        `json:"payment_method,omitempty"`
	ShippingAddr  string        `json:"shipping_address,omitempty"`
	Items         []OrderItemInfo `json:"items"`
	CreatedAt     string        `json:"created_at"`
}

type OrderItemInfo struct {
	ID        uint    `json:"id"`
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type GetOrdersReq struct {
	Page     int `json:"page,default=1"`
	PageSize int `json:"page_size,default=10"`
}
