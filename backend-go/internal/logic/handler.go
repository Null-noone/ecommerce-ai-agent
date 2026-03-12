package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"ecommerce-ai-agent/internal/svc"
	"ecommerce-ai-agent/internal/types"
	"ecommerce-ai-agent/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/threading"
	"github.com/zeromicro/go-zero/core/trace"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ==================== Auth Handlers ====================

func Register(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			types.BadRequest(w, "Invalid request body")
			return
		}

		// Validate input
		if req.Username == "" || req.Password == "" {
			types.BadRequest(w, "Username and password are required")
			return
		}

		// Check if user exists
		var existingUser model.User
		err := svcCtx.DB.Where("username = ?", req.Username).First(&existingUser).Error
		if err == nil {
			types.BadRequest(w, "Username already exists")
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			types.ErrorMsg(w, "Database error")
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			types.ErrorMsg(w, "Failed to process password")
			return
		}

		// Create user
		user := &model.User{
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			Email:        req.Email,
		}

		if err := svcCtx.DB.Create(user).Error; err != nil {
			types.ErrorMsg(w, "Failed to create user")
			return
		}

		types.SuccessMsg(w, "User registered successfully")
	}
}

func Login(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			types.BadRequest(w, "Invalid request body")
			return
		}

		// Validate input
		if req.Username == "" || req.Password == "" {
			types.BadRequest(w, "Username and password are required")
			return
		}

		// Find user
		var user model.User
		err := svcCtx.DB.Where("username = ?", req.Username).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			types.Unauthorized(w, "Invalid username or password")
			return
		}
		if err != nil {
			types.ErrorMsg(w, "Database error")
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			types.Unauthorized(w, "Invalid username or password")
			return
		}

		// Generate JWT
		now := time.Now()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  user.ID,
			"username": user.Username,
			"exp":      now.Add(time.Hour * 24 * time.Duration(svcCtx.Config.Auth.Expire)).Unix(),
			"iat":      now.Unix(),
		})

		tokenString, err := token.SignedString([]byte(svcCtx.Config.Auth.Secret))
		if err != nil {
			types.ErrorMsg(w, "Failed to generate token")
			return
		}

		types.Success(w, types.LoginResp{
			Token: tokenString,
			User: types.UserInfo{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
			},
		})
	}
}

// ==================== Product Handlers ====================

func GetProducts(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pageSize := r.URL.Query().Get("page_size")

		var pageNum, size int = 1, 10
		fmt.Sscanf(page, "%d", &pageNum)
		fmt.Sscanf(pageSize, "%d", &size)

		// Limit page size
		if size > 50 {
			size = 50
		}

		var products []model.Product
		var total int64

		offset := (pageNum - 1) * size
		err := svcCtx.DB.Model(&model.Product{}).Count(&total).Error
		if err != nil {
			types.ErrorMsg(w, "Database error")
			return
		}

		err = svcCtx.DB.Offset(offset).Limit(size).Find(&products).Error
		if err != nil {
			types.ErrorMsg(w, "Failed to fetch products")
			return
		}

		// Convert to response type
		productInfos := make([]types.ProductInfo, len(products))
		for i, p := range products {
			productInfos[i] = convertProduct(p)
		}

		types.Success(w, map[string]interface{}{
			"products": productInfos,
			"total":    total,
			"page":     pageNum,
			"page_size": size,
		})
	}
}

func GetProduct(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var product model.Product

		var productID uint
		fmt.Sscanf(id, "%d", &productID)

		if err := svcCtx.DB.First(&product, productID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			types.NotFound(w, "Product not found")
			return
		}
		if err != nil {
			types.ErrorMsg(w, "Database error")
			return
		}

		types.Success(w, convertProduct(product))
	}
}

// ==================== Search Handlers ====================

func SemanticSearch(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit := r.URL.Query().Get("limit")

		if query == "" {
			types.BadRequest(w, "Query parameter 'q' is required")
			return
		}

		var topK int = 10
		fmt.Sscanf(limit, "%d", &topK)
		if topK > 50 {
			topK = 50
		}

		// Call Python AI service
		pythonURL := "http://python-svc:8000/agent/semantic_search"
		
		reqBody, _ := json.Marshal(map[string]interface{}{
			"query": query,
			"top_k": topK,
		})

		resp, err := http.Post(pythonURL, "application/json", nil)
		
		var products []model.Product
		fallback := false

		if err != nil {
			// Fallback to basic SQL search
			fallback = true
			likeQuery := "%" + query + "%"
			svcCtx.DB.Where("name LIKE ? OR description LIKE ?", likeQuery, likeQuery).Limit(topK).Find(&products)
		} else {
			defer resp.Body.Close()

			var searchResult map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
				types.ErrorMsg(w, "Failed to parse search results")
				return
			}

			productIDs, _ := searchResult["product_ids"].([]interface{})
			
			var ids []uint
			for _, id := range productIDs {
				ids = append(ids, uint(id.(float64)))
			}

			if len(ids) > 0 {
				svcCtx.DB.Where("id IN ?", ids).Find(&products)
			}
		}

		// Convert to response type
		productInfos := make([]types.ProductInfo, len(products))
		for i, p := range products {
			productInfos[i] = convertProduct(p)
		}

		result := map[string]interface{}{
			"products": productInfos,
			"query":    query,
			"total":    len(products),
		}
		
		if fallback {
			result["fallback"] = true
			result["message"] = "Using basic search (AI service unavailable)"
		}

		types.Success(w, result)
	}
}

// ==================== Order Handlers ====================

func CreateOrder(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		userID := r.Context().Value("user_id")
		if userID == nil {
			types.Unauthorized(w, "Please login first")
			return
		}

		var req types.CreateOrderReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			types.BadRequest(w, "Invalid request body")
			return
		}

		if len(req.Items) == 0 {
			types.BadRequest(w, "Order items cannot be empty")
			return
		}

		// Use distributed lock to prevent overselling
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		lockKey := fmt.Sprintf("order:%d", userID.(uint))
		lock := svcCtx.NewLock(lockKey, 10)

		err := lock.LockWithFunc(ctx, func() error {
			// Start transaction
			tx := svcCtx.DB.Begin()

			var totalAmount float64
			var orderItems []model.OrderItem

			for _, item := range req.Items {
				if item.Quantity <= 0 {
					tx.Rollback()
					return errors.New("invalid quantity")
				}

				var product model.Product
				if err := tx.First(&product, item.ProductID).Error; err != nil {
					tx.Rollback()
					return errors.New(fmt.Sprintf("product not found: %d", item.ProductID))
				}

				if product.Stock < item.Quantity {
					tx.Rollback()
					return errors.New(fmt.Sprintf("insufficient stock for product: %s", product.Name))
				}

				// Deduct stock
				if err := tx.Model(&product).Update("stock", product.Stock-item.Quantity).Error; err != nil {
					tx.Rollback()
					return err
				}

				orderItems = append(orderItems, model.OrderItem{
					ProductID: item.ProductID,
					Quantity:  item.Quantity,
					Price:     product.Price,
				})

				totalAmount += product.Price * float64(item.Quantity)
			}

			// Create order
			order := &model.Order{
				UserID:      userID.(uint),
				TotalAmount: totalAmount,
				Status:      "pending",
			}

			if err := tx.Create(order).Error; err != nil {
				tx.Rollback()
				return err
			}

			// Create order items
			for i := range orderItems {
				orderItems[i].OrderID = order.ID
			}
			if err := tx.Create(&orderItems).Error; err != nil {
				tx.Rollback()
				return err
			}

			// Commit transaction
			if err := tx.Commit().Error; err != nil {
				return err
			}

			// Send Kafka message asynchronously
			threading.GoSafe(func() {
				svcCtx.KafkaProducer.SendOrderCreatedEvent(order.ID, order.UserID, order.TotalAmount)
			})

			// Return order
			types.Success(w, convertOrder(order, orderItems))
			return nil
		})

		if err != nil {
			types.BadRequest(w, err.Error())
		}
	}
}

func GetOrders(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		if userID == nil {
			types.Unauthorized(w, "Please login first")
			return
		}

		page := r.URL.Query().Get("page")
		pageSize := r.URL.Query().Get("page_size")

		var pageNum, size int = 1, 10
		fmt.Sscanf(page, "%d", &pageNum)
		fmt.Sscanf(pageSize, "%d", &size)

		if size > 50 {
			size = 50
		}

		var orders []model.Order
		var total int64

		offset := (pageNum - 1) * size
		err := svcCtx.DB.Model(&model.Order{}).Where("user_id = ?", userID).Count(&total).Error
		if err != nil {
			types.ErrorMsg(w, "Database error")
			return
		}

		err = svcCtx.DB.Preload("Items").Where("user_id = ?", userID).Offset(offset).Limit(size).Find(&orders).Error
		if err != nil {
			types.ErrorMsg(w, "Failed to fetch orders")
			return
		}

		// Convert to response type
		orderInfos := make([]types.OrderInfo, len(orders))
		for i, o := range orders {
			orderInfos[i] = convertOrder(o, o.Items)
		}

		types.Success(w, map[string]interface{}{
			"orders":    orderInfos,
			"total":     total,
			"page":      pageNum,
			"page_size": size,
		})
	}
}

// ==================== Helpers ====================

func convertProduct(p model.Product) types.ProductInfo {
	return types.ProductInfo{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		ImageURL:    p.ImageURL,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func convertOrder(o model.Order, items []model.OrderItem) types.OrderInfo {
	itemInfos := make([]types.OrderItemInfo, len(items))
	for i, item := range items {
		itemInfos[i] = types.OrderItemInfo{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return types.OrderInfo{
		ID:          o.ID,
		UserID:      o.UserID,
		TotalAmount: o.TotalAmount,
		Status:      o.Status,
		Items:       itemInfos,
		CreatedAt:   o.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
