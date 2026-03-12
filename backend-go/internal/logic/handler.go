package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ecommerce-ai-agent/internal/svc"
	"ecommerce-ai-agent/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/core/threading"
	"github.com/zeromicro/go-zero/rest/httpx"
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

func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		user := &model.User{
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			Email:        req.Email,
		}

		if err := svcCtx.DB.Create(user).Error; err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJson(w, map[string]interface{}{
			"message": "User registered successfully",
			"user_id": user.ID,
		})
	}
}

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		var user model.User
		if err := svcCtx.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
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
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJson(w, map[string]interface{}{
			"token": tokenString,
			"user": user,
		})
	}
}

func ListProductsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var products []model.Product
		svcCtx.DB.Find(&products)
		httpx.OkJson(w, products)
	}
}

func GetProductHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var product model.Product
		if err := svcCtx.DB.First(&product, id).Error; err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJson(w, product)
	}
}

func SemanticSearchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		page := r.URL.Query().Get("page")
		limit := r.URL.Query().Get("limit")

		if query == "" {
			httpx.ErrorCtx(r.Context(), w, fmt.Errorf("query parameter 'q' is required"))
			return
		}

		if page == "" {
			page = "1"
		}
		if limit == "" {
			limit = "10"
		}

		// Call Python service for semantic search
		pythonURL := fmt.Sprintf("%s/agent/semantic_search", svcCtx.Config.PythonService.Endpoints[0])
		
		reqBody, _ := json.Marshal(map[string]interface{}{
			"query": query,
			"top_k": limit,
		})

		resp, err := http.Post(pythonURL, "application/json", nil)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, fmt.Errorf("failed to call AI service: %v", err))
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

		httpx.OkJson(w, map[string]interface{}{
			"products": products,
			"query":    query,
		})
	}
}

func CreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (set by auth middleware)
		userID := r.Context().Value("user_id")
		if userID == nil {
			httpx.ErrorCtx(r.Context(), w, fmt.Errorf("unauthorized"))
			return
		}

		var req CreateOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// Start transaction
		tx := svcCtx.DB.Begin()

		var totalAmount float64
		var orderItems []model.OrderItem

		for _, item := range req.Items {
			var product model.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				tx.Rollback()
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("product not found: %d", item.ProductID))
				return
			}

			if product.Stock < item.Quantity {
				tx.Rollback()
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("insufficient stock for product: %s", product.Name))
				return
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
			httpx.ErrorCtx(r.Context(), w, err)
			return
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
			msg := map[string]interface{}{
				"event":       "order_created",
				"order_id":    order.ID,
				"user_id":     order.UserID,
				"total_amount": order.TotalAmount,
				"timestamp":   time.Now().Unix(),
			}
			msgBytes, _ := json.Marshal(msg)
			svcCtx.KafkaProducer.SendMessage(svcCtx.Config.Kafka.Topic, fmt.Sprintf("%d", order.ID), msgBytes)
		})

		httpx.OkJson(w, order)
	}
}

func ListOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id")
		if userID == nil {
			httpx.ErrorCtx(r.Context(), w, fmt.Errorf("unauthorized"))
			return
		}

		var orders []model.Order
		svcCtx.DB.Preload("Items").Where("user_id = ?", userID).Find(&orders)
		httpx.OkJson(w, orders)
	}
}
