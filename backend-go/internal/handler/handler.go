package handler

import (
	"net/http"

	"ecommerce-ai-agent/internal/logic"
	"ecommerce-ai-agent/internal/middleware"
	"ecommerce-ai-agent/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	
	// Add middlewares
	server.Use(middleware.CORS("*"))
	
	// Auth routes (public)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/auth/register",
				Handler: logic.Register(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/auth/login",
				Handler: logic.Login(serverCtx),
			},
		},
	)

	// Product routes (public)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/products",
				Handler: logic.GetProducts(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/products/:id",
				Handler: logic.GetProduct(serverCtx),
			},
		},
	)

	// Search routes (public)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/search/semantic",
				Handler: logic.SemanticSearch(serverCtx),
			},
		},
	)

	// Order routes (protected by JWT)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/orders",
				Handler: middleware.JWTAuth(serverCtx.Config.Auth.Secret)(http.HandlerFunc(logic.CreateOrder(serverCtx))),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/orders",
				Handler: middleware.JWTAuth(serverCtx.Config.Auth.Secret)(http.HandlerFunc(logic.GetOrders(serverCtx))),
			},
		},
	)
}
