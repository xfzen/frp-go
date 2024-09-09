package fmgr

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"frpgo/client"
	gconfig "frpgo/config"
	"frpgo/fmgr/webhook"
	"frpgo/pkg/config"
	"frpgo/pkg/util/log"
	"frpgo/pkg2/utils2"

	"github.com/zeromicro/go-zero/core/logx"
)

func CreateService(c gconfig.Config) (*client.Service, error) {
	logx.Debugf("CreateService Frp Conf: %v", utils2.PrettyJson(c.Frp))

	// setup webhook
	webhook.Setup(c)

	cfg, _, _, _, err := config.LoadClientConfig(c.Frp.Conf, true)
	if err != nil {
		return nil, err
	}

	log.InitLogger(cfg.Log.To, cfg.Log.Level, int(cfg.Log.MaxDays), cfg.Log.DisablePrintColor)

	svr, err := client.NewService(client.ServiceOptions{
		Common:         cfg,
		ProxyCfgs:      nil,
		VisitorCfgs:    nil,
		ConfigFilePath: "",
	})
	if err != nil {
		return nil, err
	}

	shouldGracefulClose := cfg.Transport.Protocol == "kcp" || cfg.Transport.Protocol == "quic"

	// Capture the exit signal if we use kcp or quic.
	if shouldGracefulClose {
		go handleTermSignal(svr)
	}

	// start service
	go svr.Run(context.Background())

	return svr, nil
}

func handleTermSignal(svr *client.Service) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	svr.GracefulClose(500 * time.Millisecond)
}
