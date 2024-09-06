package main

import (
	"flag"
	"fmt"

	"frpgo/api/internal/handler"
	"frpgo/api/internal/svc"
	"frpgo/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/frpgo-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("\nStarting server at %s:%d...\n\n", c.Host, c.Port)
	server.Start()
}
