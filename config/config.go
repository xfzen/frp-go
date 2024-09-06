package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	// Frp:
	// Conf: ./conf/frpc.toml
	// LocalIP: localhost
	// LocalPort: 8080
	// RemotePort: 9080
	Frp struct {
		Conf       string
		LocalIP    string
		LocalPort  int
		RemotePort int
	}
}
