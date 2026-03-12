package handler

import (
	"net/http"
	"ecommerce-ai-agent/internal/logic"
	"ecommerce-ai-agent/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	
	// Product routes
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/products",
				Handler: logic.ListProductsHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/products/:id",
				Handler: logic.GetProductHandler(serverCtx),
			},
		},
	)

	// Search routes
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/search/semantic",
				Handler: logic.SemanticSearchHandler(serverCtx),
			},
		},
	)

	// Order routes (require auth)
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/orders",
				Handler: logic.CreateOrderHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/api/v1/orders",
				Handler: logic.ListOrdersHandler(serverCtx),
			},
		},
	)

	// Auth routes
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/auth/register",
				Handler: logic.RegisterHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/api/v1/auth/login",
				Handler: logic.LoginHandler(serverCtx),
			},
		},
	)
}
