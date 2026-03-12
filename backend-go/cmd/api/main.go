package main

import (
	"flag"
	"fmt"

	"ecommerce-ai-agent/internal/config"
	"ecommerce-ai-agent/internal/handler"
	"ecommerce-ai-agent/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/ecommerce-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, svc.NewServiceContext())

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
