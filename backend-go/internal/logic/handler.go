package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"ecommerce-ai-agent/internal/svc"
	"ecommerce-ai-agent/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/threading"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type CreateOrderRequest struct {
	Items []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

// ==================== Auth Handlers ====================

func Register(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := &model.User{
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			Email:        req.Email,
		}

		if err := svcCtx.DB.Create(user).Error; err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User registered successfully",
			"user_id": user.ID,
		})
	}
}

func Login(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var user model.User
		if err := svcCtx.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		// Generate JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  user.ID,
			"username": user.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, err := token.SignedString([]byte(svcCtx.Config.Auth.Secret))
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": tokenString,
			"user": map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		})
	}
}

// ==================== Product Handlers ====================

func GetProducts(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var products []model.Product
		svcCtx.DB.Find(&products)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	}
}

func GetProduct(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var product model.Product
		if err := svcCtx.DB.First(&product, id).Error; err != nil {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	}
}

// ==================== Search Handlers ====================

func SemanticSearch(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit := r.URL.Query().Get("limit")

		if query == "" {
			http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
			return
		}

		if limit == "" {
			limit = "10"
		}

		// Call Python service for semantic search
		pythonURL := "http://python-svc:8000/agent/semantic_search"
		
		reqBody, _ := json.Marshal(map[string]interface{}{
			"query": query,
			"top_k": limit,
		})

		resp, err := http.Post(pythonURL, "application/json", nil)
		if err != nil {
			// Fallback to basic search
			var products []model.Product
			likeQuery := "%" + query + "%"
			svcCtx.DB.Where("name LIKE ? OR description LIKE ?", likeQuery, likeQuery).Find(&products)
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"products": products,
				"query":    query,
				"fallback": true,
			})
			return
		}
		defer resp.Body.Close()

		var searchResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&searchResult)

		productIDs, _ := searchResult["product_ids"].([]interface{})
		
		var ids []uint
		for _, id := range productIDs {
			ids = append(ids, uint(id.(float64)))
		}

		var products []model.Product
		if len(ids) > 0 {
			svcCtx.DB.Where("id IN ?", ids).Find(&products)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"products": products,
			"query":    query,
		})
	}
}

// ==================== Order Handlers ====================

func CreateOrder(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (set by auth middleware)
		userID := r.Context().Value("user_id")
		if userID == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreateOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Use distributed lock to prevent overselling
		ctx := context.Background()
		lockKey := fmt.Sprintf("order:%d", userID.(uint))
		lock := redis.NewDistributedLock(svcCtx.Redis, lockKey, 10*1000*1000*1000)

		err := lock.LockWithFunc(ctx, func() error {
			// Start transaction
			tx := svcCtx.DB.Begin()

			var totalAmount float64
			var orderItems []model.OrderItem

			for _, item := range req.Items {
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
				tx.Model(&product).Update("stock", product.Stock-item.Quantity)

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
			tx.Create(&orderItems)

			// Commit transaction
			tx.Commit()

			// Send Kafka message asynchronously
			threading.GoSafe(func() {
				svcCtx.KafkaProducer.SendOrderCreatedEvent(order.ID, order.UserID, order.TotalAmount)
			})

			// Return order
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(order)
			return nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func GetOrders(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		if userID == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var orders []model.Order
		svcCtx.DB.Preload("Items").Where("user_id = ?", userID).Find(&orders)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	}
}
