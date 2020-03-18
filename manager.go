package main

import (
	"io"
	"sync"

	"github.com/baetyl/baetyl-go/log"
	"google.golang.org/grpc"
)

// Manager Manager
type Manager interface {
	GetGRPCConnection(string, bool) (*grpc.ClientConn, error)
	io.Closer
}

// Manager is a wrapper around gRPC connection pooling
type manager struct {
	log            *log.Logger
	lock           *sync.Mutex
	connectionPool map[string]*grpc.ClientConn
}

// NewGRPCManager returns a new grpc manager
func NewGRPCManager() Manager {
	return &manager{
		log:            log.With(log.Any("main", "manager")),
		lock:           &sync.Mutex{},
		connectionPool: map[string]*grpc.ClientConn{},
	}
}

// GetGRPCConnection returns a new grpc connection for a given address and inits one if doesn't exist
func (g *manager) GetGRPCConnection(address string, recreateIfExists bool) (*grpc.ClientConn, error) {
	if val, ok := g.connectionPool[address]; ok && !recreateIfExists {
		return val, nil
	}

	g.lock.Lock()
	if val, ok := g.connectionPool[address]; ok && !recreateIfExists {
		g.lock.Unlock()
		return val, nil
	}

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		g.log.Error("failed to create connection to server", log.Error(err), log.Any("address", address))
		g.lock.Unlock()
		return nil, err
	}

	g.connectionPool[address] = conn
	g.lock.Unlock()

	return conn, nil
}

func (g *manager) Close() error {
	for address, conn := range g.connectionPool {
		err := conn.Close()
		if err != nil {
			g.log.Error("failed to close connection", log.Error(err), log.Any("address", address))
		}
	}
	return nil
}
