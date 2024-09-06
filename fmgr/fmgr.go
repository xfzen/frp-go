package fmgr

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"frpgo/client"
	gconfig "frpgo/config"
	"frpgo/pkg/config"
	v1 "frpgo/pkg/config/v1"
	"frpgo/pkg/config/v1/validation"
	"frpgo/pkg/util/log"
	"frpgo/pkg2/utils2"

	"github.com/zeromicro/go-zero/core/logx"
)

func Init(cfgfile string) {
	logx.Debugf("session Iint() conf: %v", cfgfile)
	go StartClient(cfgfile)
}

func CreateService(c gconfig.Config) (*client.Service, error) {
	logx.Debugf("CreateService Frp Conf: %v", utils2.PrettyJson(c.Frp))

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

func StartClient(path string) error {
	cfg, proxyCfgs, visitorCfgs, isLegacyFormat, err := config.LoadClientConfig(path, true)
	if err != nil {
		return err
	}
	if isLegacyFormat {
		fmt.Printf("WARNING: ini format is deprecated and the support will be removed in the future, " +
			"please use yaml/json/toml format instead!\n")
	}

	warning, err := validation.ValidateAllClientConfig(cfg, proxyCfgs, visitorCfgs)
	if warning != nil {
		fmt.Printf("WARNING: %v\n", warning)
	}
	if err != nil {
		return err
	}
	return startService(cfg, proxyCfgs, visitorCfgs, path)
}

func startService(
	cfg *v1.ClientCommonConfig,
	proxyCfgs []v1.ProxyConfigurer,
	visitorCfgs []v1.VisitorConfigurer,
	cfgFile string,
) error {
	log.InitLogger(cfg.Log.To, cfg.Log.Level, int(cfg.Log.MaxDays), cfg.Log.DisablePrintColor)

	if cfgFile != "" {
		log.Infof("start frpc service for config file [%s]", cfgFile)
		defer log.Infof("frpc service for config file [%s] stopped", cfgFile)
	}
	svr, err := client.NewService(client.ServiceOptions{
		Common:         cfg,
		ProxyCfgs:      nil,
		VisitorCfgs:    nil,
		ConfigFilePath: "",
	})
	if err != nil {
		return err
	}

	shouldGracefulClose := cfg.Transport.Protocol == "kcp" || cfg.Transport.Protocol == "quic"
	// Capture the exit signal if we use kcp or quic.
	if shouldGracefulClose {
		go handleTermSignal(svr)
	}
	return svr.Run(context.Background())
}

func handleTermSignal(svr *client.Service) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	svr.GracefulClose(500 * time.Millisecond)
}
