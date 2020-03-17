package main

import (
	"google.golang.org/grpc"
	"sync"
)

// Manager is a wrapper around gRPC connection pooling
type Manager struct {
	lock           *sync.Mutex
	connectionPool map[string]*grpc.ClientConn
}

// NewGRPCManager returns a new grpc manager
func NewGRPCManager() *Manager {
	return &Manager{
		lock:           &sync.Mutex{},
		connectionPool: map[string]*grpc.ClientConn{},
	}
}

// GetGRPCConnection returns a new grpc connection for a given address and inits one if doesn't exist
func (g *Manager) GetGRPCConnection(address string, recreateIfExists bool) (*grpc.ClientConn, error) {
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
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		g.lock.Unlock()
		return nil, err
	}

	g.connectionPool[address] = conn
	g.lock.Unlock()

	return conn, nil
}
