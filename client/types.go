package client

import (
	"context"
	"net"
	"sync"
	"time"

	"frpgo/client/proxy"
	"frpgo/pkg/auth"
	v1 "frpgo/pkg/config/v1"
	"frpgo/pkg/msg"
	httppkg "frpgo/pkg/util/http"
)

// Service is the client service that connects to frps and provides proxy services.
type Service struct {
	ctlMu sync.RWMutex
	// manager control connection with server
	ctl *Control
	// Uniq id got from frps, it will be attached to loginMsg.
	runID string

	// Sets authentication based on selected method
	authSetter auth.Setter

	// web server for admin UI and apis
	webServer *httppkg.Server

	cfgMu       sync.RWMutex
	common      *v1.ClientCommonConfig
	proxyCfgs   []v1.ProxyConfigurer
	visitorCfgs []v1.VisitorConfigurer
	clientSpec  *msg.ClientSpec

	// The configuration file used to initialize this client, or an empty
	// string if no configuration file was used.
	configFilePath string

	// service context
	ctx context.Context
	// call cancel to stop service
	cancel                   context.CancelCauseFunc
	gracefulShutdownDuration time.Duration

	connectorCreator func(context.Context, *v1.ClientCommonConfig) Connector
	handleWorkConnCb func(*v1.ProxyBaseConfig, net.Conn, *msg.StartWorkConn) bool
}

// ServiceOptions contains options for creating a new client service.
type ServiceOptions struct {
	Common      *v1.ClientCommonConfig
	ProxyCfgs   []v1.ProxyConfigurer
	VisitorCfgs []v1.VisitorConfigurer

	// ConfigFilePath is the path to the configuration file used to initialize.
	// If it is empty, it means that the configuration file is not used for initialization.
	// It may be initialized using command line parameters or called directly.
	ConfigFilePath string

	// ClientSpec is the client specification that control the client behavior.
	ClientSpec *msg.ClientSpec

	// ConnectorCreator is a function that creates a new connector to make connections to the server.
	// The Connector shields the underlying connection details, whether it is through TCP or QUIC connection,
	// and regardless of whether multiplexing is used.
	//
	// If it is not set, the default frpc connector will be used.
	// By using a custom Connector, it can be used to implement a VirtualClient, which connects to frps
	// through a pipe instead of a real physical connection.
	ConnectorCreator func(context.Context, *v1.ClientCommonConfig) Connector

	// HandleWorkConnCb is a callback function that is called when a new work connection is created.
	//
	// If it is not set, the default frpc implementation will be used.
	HandleWorkConnCb func(*v1.ProxyBaseConfig, net.Conn, *msg.StartWorkConn) bool
}

type StatusExporter interface {
	GetProxyStatus(name string) (*proxy.WorkingStatus, bool)
}

type statusExporterImpl struct {
	getProxyStatusFunc func(name string) (*proxy.WorkingStatus, bool)
}

func (s *statusExporterImpl) GetProxyStatus(name string) (*proxy.WorkingStatus, bool) {
	return s.getProxyStatusFunc(name)
}
