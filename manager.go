package main

import (
	"crypto/tls"
	"io"
	"sync"

	"github.com/baetyl/baetyl-go/log"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
)

// Manager Manager
type Manager interface {
	GetGRPCConnection(string, bool) (*grpc.ClientConn, error)
	GetHttpClient() *fasthttp.Client
	io.Closer
}

type manager struct {
	log            *log.Logger
	cfg            *ClientConfig
	lock           *sync.Mutex
	httpClient     *fasthttp.Client
	connectionPool map[string]*grpc.ClientConn
}

// NewGRPCManager
func NewManager(cfg ClientConfig) Manager {
	httpClient := &fasthttp.Client{
		MaxConnsPerHost: cfg.Http.MaxConnsPerHost,
		// TODO: support tls
		TLSConfig:                 &tls.Config{InsecureSkipVerify: true},
		ReadTimeout:               cfg.Http.ReadTimeout,
		MaxIdemponentCallAttempts: cfg.Http.MaxIdemponentCallAttempts,
		MaxConnDuration:           cfg.Http.MaxConnDuration,
	}
	return &manager{
		log:            log.With(log.Any("main", "manager")),
		cfg:            &cfg,
		lock:           &sync.Mutex{},
		httpClient:     httpClient,
		connectionPool: map[string]*grpc.ClientConn{},
	}
}

// GetGRPCConnection returns a new grpc connection for a given address and inits one if doesn't exist
func (g *manager) GetGRPCConnection(address string, recreateIfExists bool) (*grpc.ClientConn, error) {
	if val, ok := g.connectionPool[address]; ok && !recreateIfExists {
		return val, nil
	}

	g.lock.Lock()
	defer g.lock.Unlock()
	if val, ok := g.connectionPool[address]; ok && !recreateIfExists {
		return val, nil
	}

	opts := []grpc.DialOption{
		// TODO: tls support
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		g.log.Error("failed to create connection to server", log.Error(err), log.Any("address", address))
		return nil, err
	}

	g.connectionPool[address] = conn
	return conn, nil
}

// GetHttpClient returns a fasthttp client
func (g *manager) GetHttpClient() *fasthttp.Client {
	return g.httpClient
}

func (g *manager) Close() error {
	for address, conn := range g.connectionPool {
		err := conn.Close()
		if err != nil {
			g.log.Warn("failed to close connection", log.Error(err), log.Any("address", address))
		}
	}
	return nil
}
