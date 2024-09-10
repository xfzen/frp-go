// Copyright 2023 The frp Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxy

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sync"

	"github.com/samber/lo"
	"github.com/zeromicro/go-zero/core/logx"

	"frpgo/client/event"
	"frpgo/fmgr/webhook"
	v1 "frpgo/pkg/config/v1"
	"frpgo/pkg/msg"
	"frpgo/pkg/transport"
	"frpgo/pkg/util/xlog"
	"frpgo/pkg2/utils2"
)

type Manager struct {
	proxies            map[string]*Wrapper
	msgTransporter     transport.MessageTransporter
	inWorkConnCallback func(*v1.ProxyBaseConfig, net.Conn, *msg.StartWorkConn) bool

	closed bool
	mu     sync.RWMutex

	clientCfg *v1.ClientCommonConfig

	ctx context.Context
}

func NewManager(
	ctx context.Context,
	clientCfg *v1.ClientCommonConfig,
	msgTransporter transport.MessageTransporter,
) *Manager {
	return &Manager{
		proxies:        make(map[string]*Wrapper),
		msgTransporter: msgTransporter,
		closed:         false,
		clientCfg:      clientCfg,
		ctx:            ctx,
	}
}

func (pm *Manager) StartProxy(name string, remoteAddr string, serverRespErr string) error {
	logx.Debugf("StartProxy name: %v, remoteAddr: %v, serverRespErr: %v", name, remoteAddr, serverRespErr)

	pm.mu.RLock()
	pxy, ok := pm.proxies[name]
	pm.mu.RUnlock()
	if !ok {
		return fmt.Errorf("proxy [%s] not found", name)
	}

	err := pxy.SetRunningStatus(remoteAddr, serverRespErr)
	if err != nil {
		return err
	}
	return nil
}

func (pm *Manager) SetInWorkConnCallback(cb func(*v1.ProxyBaseConfig, net.Conn, *msg.StartWorkConn) bool) {
	pm.inWorkConnCallback = cb
}

func (pm *Manager) Close() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, pxy := range pm.proxies {
		pxy.Stop()
	}
	pm.proxies = make(map[string]*Wrapper)
}

func (pm *Manager) HandleWorkConn(name string, workConn net.Conn, m *msg.StartWorkConn) {
	logx.Debugf("HandleWorkConn")

	pm.mu.RLock()
	pw, ok := pm.proxies[name]
	pm.mu.RUnlock()
	if ok {
		pw.InWorkConn(workConn, m)
	} else {
		workConn.Close()
	}
}

func (pm *Manager) HandleEvent(payload interface{}) error {
	var m msg.Message
	switch e := payload.(type) {
	case *event.StartProxyPayload:
		m = e.NewProxyMsg
	case *event.CloseProxyPayload:
		m = e.CloseProxyMsg
	default:
		return event.ErrPayloadType
	}

	return pm.msgTransporter.Send(m)
}

func (pm *Manager) GetAllProxyStatus() []*WorkingStatus {
	ps := make([]*WorkingStatus, 0)
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for _, pxy := range pm.proxies {
		ps = append(ps, pxy.GetStatus())
	}
	return ps
}

func (pm *Manager) GetProxyStatus(name string) (*WorkingStatus, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if pxy, ok := pm.proxies[name]; ok {
		return pxy.GetStatus(), true
	}
	return nil, false
}

func (pm *Manager) UpdateAll(proxyCfgs []v1.ProxyConfigurer) {
	xl := xlog.FromContextSafe(pm.ctx)
	proxyCfgsMap := lo.KeyBy(proxyCfgs, func(c v1.ProxyConfigurer) string {
		return c.GetBaseConfig().Name
	})
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delPxyNames := make([]string, 0)
	for name, pxy := range pm.proxies {
		del := false
		cfg, ok := proxyCfgsMap[name]
		if !ok || !reflect.DeepEqual(pxy.Cfg, cfg) {
			del = true
		}

		if del {
			delPxyNames = append(delPxyNames, name)
			delete(pm.proxies, name)
			pxy.Stop()
		}
	}
	if len(delPxyNames) > 0 {
		xl.Infof("proxy removed: %s", delPxyNames)
	}

	addPxyNames := make([]string, 0)
	for _, cfg := range proxyCfgs {
		name := cfg.GetBaseConfig().Name

		logx.Debugf("Name: %v", name)

		if _, ok := pm.proxies[name]; !ok {
			pxy := NewWrapper(pm.ctx, cfg, pm.clientCfg, pm.HandleEvent, pm.msgTransporter)
			if pm.inWorkConnCallback != nil {
				pxy.SetInWorkConnCallback(pm.inWorkConnCallback)
			}
			pm.proxies[name] = pxy
			addPxyNames = append(addPxyNames, name)

			pxy.Start()
		}
	}
	if len(addPxyNames) > 0 {
		xl.Infof("proxy added: %s", addPxyNames)
	}
}

// 创建新的代理（tcp或http）
func (pm *Manager) CreateProxy(proxyType string, name string, localIP string, localPort int, remotePort int) error {
	cfg := v1.NewProxyConfigurerByType(v1.ProxyType(proxyType))
	if cfg == nil {
		return fmt.Errorf("new proxy configurer error")
	}

	cfg.GetBaseConfig().Name = name
	cfg.GetBaseConfig().Type = proxyType
	// cfg.GetBaseConfig().LocalIP = localIP
	cfg.GetBaseConfig().LocalPort = localPort

	serverIp := pm.clientCfg.ServerAddr

	switch proxyType {
	case string(v1.ProxyTypeTCP):
		cfg.(*v1.TCPProxyConfig).LocalIP = localIP
		cfg.(*v1.TCPProxyConfig).LocalPort = localPort
		cfg.(*v1.TCPProxyConfig).RemotePort = remotePort

	case string(v1.ProxyTypeHTTP):
		cfg.(*v1.HTTPProxyConfig).LocalIP = localIP
		cfg.(*v1.HTTPProxyConfig).LocalPort = localPort

	default:
		cfg.(*v1.HTTPProxyConfig).LocalIP = localIP
	}

	logx.Debugf("CreateProxy serverIp: %v, proxyCfg: %v", serverIp, utils2.PrettyJson(cfg))

	pxy := NewWrapper(pm.ctx, cfg, pm.clientCfg, pm.HandleEvent, pm.msgTransporter)
	if pm.inWorkConnCallback != nil {
		pxy.SetInWorkConnCallback(pm.inWorkConnCallback)
	}

	pm.proxies[name] = pxy

	pxy.Start()

	return nil
}

func (pm *Manager) GetProxyDetail(name string) (*WorkingDetial, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if pxy, ok := pm.proxies[name]; ok {
		return pxy.GetDetial(), true
	}
	return nil, false
}

// 代理是否已创建
func (pm *Manager) IsProxyExist(name string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if _, ok := pm.proxies[name]; ok {
		// webhook
		proxyDetial, isSuccess := pm.GetProxyDetail(name)
		if isSuccess {
			webhook.PushProxyDetail(proxyDetial)

			return true
		}
		return false
	}

	return false
}
