package svc

import (
	"frpgo/client"
	"frpgo/config"
	"frpgo/fmgr"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config config.Config

	ProxyService *client.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	svr, err := fmgr.CreateService(c)
	if err != nil {
		logx.Errorf("FMGR CreateServer error %v", c.Frp.Conf)
		return nil
	}

	return &ServiceContext{
		Config:       c,
		ProxyService: svr,
	}
}
