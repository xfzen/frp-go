package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	// FRPConf: ./conf/frpc.toml
	Frp struct {
		Conf string
	}

	// webhook
	Webhook struct {
		Url string
	}
}
