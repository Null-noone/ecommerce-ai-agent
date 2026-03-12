package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ecommerce-ai-agent/internal/config"
	"ecommerce-ai-agent/internal/handler"
	"ecommerce-ai-agent/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/ecommerce.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// Configure logging
	logx.MustSetup(logx.LogConf{
		ServiceName: "ecommerce-api",
		Mode:        "console",
		Level:       "info",
	})

	// Create REST server
	server := rest.MustNewServer(c.RestConf,
		rest.WithCors("*"),
		rest.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized", "message": "Please login first"}`))
		}),
		rest.WithMiddleware(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Log request
				logx.Infof("%s %s", r.Method, r.URL.Path)
				next.ServeHTTP(w, r)
			})
		}),
	)
	defer server.Stop()

	// Initialize service context
	serviceContext := svc.NewServiceContext(&c)

	// Register handlers
	handler.RegisterHandlers(server, serviceContext)

	// Start graceful shutdown
	go func() {
		fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
		server.Start()
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logx.Info("Shutting down server...")
}
