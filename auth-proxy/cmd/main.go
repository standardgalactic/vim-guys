package main

import (
	"github.com/joho/godotenv"
	"vim-guys.theprimeagen.tv/pkg/config"
	"vim-guys.theprimeagen.tv/pkg/proxy"
	"vim-guys.theprimeagen.tv/pkg/server"
)

func main() {
	godotenv.Load()

	ctx := config.NewAuthConfig(config.ProxyConfigParamsFromEnv())
	p := proxy.NewProxy(ctx)
	s := server.NewProxyServer()

	p.AddInterceptor(s)
	p.Start()
}
