package main

import (
	"flag"
	"fmt"
	"net/http"

	"ecommerce-ai-agent/internal/config"
	"ecommerce-ai-agent/internal/handler"
	"ecommerce-ai-agent/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/ecommerce.api", "the api file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf,
		rest.WithCors("*"),
		rest.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
		}),
	)
	defer server.Stop()

	handler.RegisterHandlers(server, svc.NewServiceContext())

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
