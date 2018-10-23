// Copyright 2018 The go-pttai Authors
// This file is part of the go-pttai library.
//
// The go-pttai library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-pttai library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-pttai library. If not, see <http://www.gnu.org/licenses/>.

// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package node

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/ailabstw/go-pttai/event"
	"github.com/ailabstw/go-pttai/internal/debug"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/rpc"
	pkgservice "github.com/ailabstw/go-pttai/service"
	"github.com/oklog/oklog/pkg/flock"
)

type Node struct {
	Config *Config

	eventmux *event.TypeMux // Event multiplexer used between the services of a stack

	instanceDirLock flock.Releaser // prevents concurrent use of instance directory

	serverConfig p2p.Config
	server       *p2p.Server

	serviceFuncs []pkgservice.ServiceConstructor        // Service constructors (in dependency order)
	services     map[reflect.Type]pkgservice.PttService // Currently running services
	rpcAPIs      []rpc.API                              // List of APIs currently provided by the node

	inprocHandler *rpc.Server // In-process RPC request handler to process the API requests

	ipcEndpoint string       // IPC endpoint to listen at (empty = IPC disabled)
	ipcListener net.Listener // IPC RPC listener socket to serve API requests
	ipcHandler  *rpc.Server  // IPC RPC request handler to process the API requests

	httpEndpoint  string       // HTTP endpoint (interface + port) to listen at (empty = HTTP disabled)
	httpWhitelist []string     // HTTP RPC modules to allow through this endpoint
	httpListener  net.Listener // HTTP RPC listener socket to server API requests
	httpHandler   *rpc.Server  // HTTP RPC request handler to process the API requests
	httpServer    *http.Server

	wsEndpoint string       // Websocket endpoint (interface + port) to listen at (empty = websocket disabled)
	wsListener net.Listener // Websocket RPC listener socket to server API requests
	wsHandler  *rpc.Server  // Websocket RPC request handler to process the API requests

	lock     sync.RWMutex
	StopChan chan error

	log log.Logger
}

func New(cfg *Config) (*Node, error) {
	if cfg.DataDir != "" {
		absdatadir, err := filepath.Abs(cfg.DataDir)
		if err != nil {
			return nil, err
		}
		cfg.DataDir = absdatadir
	}

	// Ensure that the instance name doesn't cause weird conflicts with
	// other files in the data directory.
	if strings.ContainsAny(cfg.Name, `/\`) {
		return nil, errors.New(`Config.Name must not contain '/' or '\'`)
	}
	if cfg.Name == DataDirDefaultKeyStore {
		return nil, errors.New(`Config.Name cannot be "` + DataDirDefaultKeyStore + `"`)
	}
	if strings.HasSuffix(cfg.Name, ".ipc") {
		return nil, errors.New(`Config.Name cannot end in ".ipc"`)
	}
	if strings.HasSuffix(cfg.Name, ".me") {
		return nil, errors.New(`Config.Name cannot end in .me`)
	}
	if strings.HasSuffix(cfg.Name, "mykey") {
		return nil, errors.New(`Config.Name cannot end in mykey`)
	}
	if strings.HasSuffix(cfg.Name, ".revoke") {
		return nil, errors.New(`Config.Name cannot end in ".revoke"`)
	}

	if cfg.Logger == nil {
		cfg.Logger = log.New()
	}

	return &Node{
		Config:       cfg,
		serviceFuncs: []pkgservice.ServiceConstructor{},
		ipcEndpoint:  cfg.IPCEndpoint(),
		httpEndpoint: cfg.HTTPEndpoint(),
		wsEndpoint:   cfg.WSEndpoint(),
		eventmux:     new(event.TypeMux),
		log:          cfg.Logger,
	}, nil
}

// Register injects a new service into the node's stack. The service created by
// the passed constructor must be unique in its type with regard to sibling ones.
func (n *Node) Register(constructor pkgservice.ServiceConstructor) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.server != nil {
		return ErrNodeRunning
	}
	n.serviceFuncs = append(n.serviceFuncs, constructor)
	return nil
}

func (n *Node) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.server != nil {
		return ErrNodeRunning
	}

	if err := n.openDataDir(); err != nil {
		return err
	}

	// Initialize the p2p server. This creates the node key and
	// discovery databases.
	n.serverConfig = n.Config.P2P
	n.serverConfig.PrivateKey = n.Config.NodeKey()
	n.serverConfig.Name = n.Config.NodeName()
	n.serverConfig.Logger = n.log
	if n.serverConfig.StaticNodes == nil {
		n.serverConfig.StaticNodes = n.Config.StaticNodes()
	}
	if n.serverConfig.TrustedNodes == nil {
		n.serverConfig.TrustedNodes = n.Config.TrustedNodes()
	}
	if n.serverConfig.NodeDatabase == "" {
		n.serverConfig.NodeDatabase = n.Config.NodeDB()
	}
	running := &p2p.Server{Config: n.serverConfig}
	n.log.Info("Starting peer-to-peer node", "instance", n.serverConfig.Name)

	// Otherwise copy and specialize the P2P configuration
	services := make(map[reflect.Type]pkgservice.PttService)
	n.log.Info("Starting", "serviceFuncs", n.serviceFuncs, "services", services)
	for _, constructor := range n.serviceFuncs {
		// Create a new context for the particular service
		ctx := &pkgservice.ServiceContext{
			Services: services,
			EventMux: n.eventmux,
		}
		// Construct and save the service
		n.log.Info("in serviceFunc-loop: to constructor", "services", services)
		service, err := constructor(ctx)
		n.log.Info("in serviceFunc-loop", "services", services, "service", service, "e", err)
		if err != nil {
			log.Error("in serviceFunc-loop: unable to constructor", "e", err)
			return err
		}
		kind := reflect.TypeOf(service)
		if _, exists := services[kind]; exists {
			log.Error("in serviceFunc-loop: dup-resource", "kind", kind)
			return &DuplicateServiceError{Kind: kind}
		}
		services[kind] = service
	}
	n.log.Info("after serviceFunc-loop", "services", services)

	// Gather the protocols and start the freshly assembled P2P server
	for _, service := range services {
		running.Protocols = append(running.Protocols, service.Protocols()...)
	}

	n.log.Info("to running.Start", "running.Protocols", running.Protocols)

	// p2p-server start
	if err := running.Start(); err != nil {
		return ConvertFileLockError(err)
	}

	// Start each of the services
	started := []reflect.Type{}
	for kind, service := range services {
		// Start the next service, stopping all previous upon failure
		n.log.Info("to service.Start", "kind", kind, "service", service)

		if err := service.Start(running); err != nil {
			n.log.Error("something went wrong with service-starting", "e", err, "service", service)

			for _, kind := range started {
				services[kind].Stop()
			}
			running.Stop()

			return err
		}
		// Mark the service started for potential cleanup
		started = append(started, kind)
	}

	// Lastly start the configured RPC interfaces
	if err := n.startRPC(services); err != nil {
		n.log.Error("something went wrong with startRPC", "e", err)
		for _, service := range services {
			service.Stop()
		}
		running.Stop()
		return err
	}

	// Finish initializing the startup
	n.services = services
	n.server = running
	n.StopChan = make(chan error)

	return nil
}

func (n *Node) Stop(isRevoke bool, isRestart bool) error {
	n.lock.Lock()
	defer n.lock.Unlock()

	// Short circuit if the node's not running
	if n.server == nil {
		return ErrNodeStopped
	}

	// Terminate the API, services and the p2p server.
	n.stopWS()
	n.stopHTTP()
	n.stopIPC()
	n.rpcAPIs = nil
	failure := &StopError{
		Services: make(map[reflect.Type]error),
	}

	// close mutex
	log.Info("to stop eventmux")
	n.eventmux.Stop()
	log.Info("after stop eventmux")

	for kind, service := range n.services {
		log.Info("to stop service", "kind", kind)
		if err := service.Stop(); err != nil {
			failure.Services[kind] = err
		}
		log.Info("after stop service", "kind", kind)
	}

	log.Info("to stop server")
	n.server.Stop()
	log.Info("after stop server")

	// set nil
	n.services = nil
	n.server = nil

	// Release instance directory lock.
	if n.instanceDirLock != nil {
		n.log.Info("to release instanceDirLock")
		if err := n.instanceDirLock.Release(); err != nil {
			n.log.Error("Can't release datadir lock", "err", err)
		}
		n.instanceDirLock = nil
	}

	switch {
	case isRevoke:
		n.Config.RevokeKeyPath()
	case isRestart:
	default:
		close(n.StopChan)
	}

	if len(failure.Services) > 0 {
		return failure
	}

	return nil
}

// Restart terminates a running node and boots up a new one in its place. If the
// node isn't running, an error is returned.
func (n *Node) Restart(isRevoke bool, isRestart bool) error {
	if err := n.Stop(isRevoke, isRestart); err != nil {
		return err
	}
	if err := n.Start(); err != nil {
		return err
	}
	return nil
}

func (n *Node) Services() map[reflect.Type]pkgservice.PttService {
	return n.services

}

// RPCHandler returns the in-process RPC request handler.
func (n *Node) RPCHandler() (*rpc.Server, error) {
	n.lock.RLock()
	defer n.lock.RUnlock()

	if n.inprocHandler == nil {
		return nil, ErrNodeStopped
	}
	return n.inprocHandler, nil
}

// Server retrieves the currently running P2P network layer. This method is meant
// only to inspect fields of the currently running server, life cycle management
// should be left to this Node entity.
func (n *Node) Server() *p2p.Server {
	n.lock.RLock()
	defer n.lock.RUnlock()

	return n.server
}

// DataDir retrieves the current datadir used by the protocol stack.
// Deprecated: No files should be stored in this directory, use InstanceDir instead.
func (n *Node) DataDir() string {
	return n.Config.DataDir
}

func (n *Node) openDataDir() error {
	log.Info("start", "DataDir", n.Config.DataDir)

	if n.Config.DataDir == "" {
		return nil // ephemeral
	}

	instdir := filepath.Join(n.Config.DataDir, n.Config.name())
	if err := os.MkdirAll(instdir, 0700); err != nil {
		return err
	}
	// Lock the instance directory to prevent concurrent use by another instance as well as
	// accidental use of the instance directory as a database.

	log.Info("to lock file", "instdir", instdir)
	release, _, err := flock.New(filepath.Join(instdir, "LOCK"))
	if err != nil {
		return ConvertFileLockError(err)
	}
	n.instanceDirLock = release
	return nil
}

// startRPC is a helper method to start all the various RPC endpoint during node
// startup. It's not meant to be called at any time afterwards as it makes certain
// assumptions about the state of the node.
func (n *Node) startRPC(services map[reflect.Type]pkgservice.PttService) error {
	// Gather all the possible APIs to surface
	apis := n.apis()
	for _, service := range services {
		apis = append(apis, service.APIs()...)
	}

	// Start the various API endpoints, terminating all in case of errors
	log.Debug("startRPC: to startInProc")
	if err := n.startInProc(apis); err != nil {
		return err
	}
	log.Debug("startRPC: to startIPC")
	if err := n.startIPC(apis); err != nil {
		n.stopInProc()
		return err
	}
	log.Debug("startRPC: to startHTTP")
	if err := n.startHTTP(n.httpEndpoint, apis, n.Config.HTTPModules, n.Config.HTTPCors, n.Config.HTTPVirtualHosts); err != nil {
		n.stopIPC()
		n.stopInProc()
		return err
	}
	log.Debug("startRPC: to startWS")
	if err := n.startWS(n.wsEndpoint, apis, n.Config.WSModules, n.Config.WSOrigins, n.Config.WSExposeAll); err != nil {
		n.stopHTTP()
		n.stopIPC()
		n.stopInProc()
		return err
	}
	// All API endpoints started successfully
	n.rpcAPIs = apis

	log.Debug("startRPC: end")
	return nil
}

// startInProc initializes an in-process RPC endpoint.
func (n *Node) startInProc(apis []rpc.API) error {
	// Register all the APIs exposed by the services
	handler := rpc.NewServer()
	for _, api := range apis {
		if err := handler.RegisterName(api.Namespace, api.Service); err != nil {
			return err
		}
		n.log.Debug("InProc registered", "service", api.Service, "namespace", api.Namespace)
	}
	n.inprocHandler = handler
	return nil
}

// stopInProc terminates the in-process RPC endpoint.
func (n *Node) stopInProc() {
	if n.inprocHandler != nil {
		n.inprocHandler.Stop()
		n.inprocHandler = nil
	}
}

// startIPC initializes and starts the IPC RPC endpoint.
func (n *Node) startIPC(apis []rpc.API) error {
	if n.ipcEndpoint == "" {
		return nil // IPC disabled.
	}

	log.Info("startIPC", "ipcEndpoint", n.ipcEndpoint)
	listener, handler, err := rpc.StartIPCEndpoint(n.ipcEndpoint, apis)
	if err != nil {
		return err
	}
	n.ipcListener = listener
	n.ipcHandler = handler
	n.log.Info("IPC endpoint opened", "url", n.ipcEndpoint)
	return nil
}

// stopIPC terminates the IPC RPC endpoint.
func (n *Node) stopIPC() {
	if n.ipcListener != nil {
		n.ipcListener.Close()
		n.ipcListener = nil

		n.log.Info("IPC endpoint closed", "endpoint", n.ipcEndpoint)
	}
	if n.ipcHandler != nil {
		n.ipcHandler.Stop()
		n.ipcHandler = nil
	}
}

// startHTTP initializes and starts the HTTP RPC endpoint.
func (n *Node) startHTTP(endpoint string, apis []rpc.API, modules []string, cors []string, vhosts []string) error {
	// Short circuit if the HTTP endpoint isn't being exposed
	n.log.Info("startHTTP", "endpoint", endpoint, "apis", apis)

	if endpoint == "" {
		return nil
	}
	listener, handler, httpServer, err := rpc.StartHTTPEndpoint(endpoint, apis, modules, cors, vhosts)
	if err != nil {
		return err
	}
	n.log.Info("HTTP endpoint opened", "url", fmt.Sprintf("http://%s", endpoint), "cors", strings.Join(cors, ","), "vhosts", strings.Join(vhosts, ","))
	// All listeners booted successfully
	n.httpEndpoint = endpoint
	n.httpListener = listener
	n.httpHandler = handler
	n.httpServer = httpServer

	return nil
}

// stopHTTP terminates the HTTP RPC endpoint.
func (n *Node) stopHTTP() {
	if n.httpListener != nil {
		n.httpListener.Close()
		n.httpListener = nil

		n.log.Info("HTTP endpoint closed", "url", fmt.Sprintf("http://%s", n.httpEndpoint))
	}
	if n.httpHandler != nil {
		n.httpHandler.Stop()
		n.httpHandler = nil
	}

	if n.httpServer != nil {
		n.httpServer.Close()
		n.httpServer = nil
	}
}

// startWS initializes and starts the websocket RPC endpoint.
func (n *Node) startWS(endpoint string, apis []rpc.API, modules []string, wsOrigins []string, exposeAll bool) error {
	// Short circuit if the WS endpoint isn't being exposed
	if endpoint == "" {
		return nil
	}
	listener, handler, err := rpc.StartWSEndpoint(endpoint, apis, modules, wsOrigins, exposeAll)
	if err != nil {
		return err
	}
	n.log.Info("WebSocket endpoint opened", "url", fmt.Sprintf("ws://%s", listener.Addr()))
	// All listeners booted successfully
	n.wsEndpoint = endpoint
	n.wsListener = listener
	n.wsHandler = handler

	return nil
}

// stopWS terminates the websocket RPC endpoint.
func (n *Node) stopWS() {
	if n.wsListener != nil {
		n.wsListener.Close()
		n.wsListener = nil

		n.log.Info("WebSocket endpoint closed", "url", fmt.Sprintf("ws://%s", n.wsEndpoint))
	}
	if n.wsHandler != nil {
		n.wsHandler.Stop()
		n.wsHandler = nil
	}
}

// apis returns the collection of RPC descriptors this node offers.
func (n *Node) apis() []rpc.API {
	return []rpc.API{
		{
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(n),
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPublicAdminAPI(n),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   debug.Handler,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(n),
			Public:    true,
		},
	}
}
