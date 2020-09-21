package function

import (
	"crypto/tls"
	"io"
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Manager Manager
type Manager interface {
	GetGRPCConnection(string, bool) (*grpc.ClientConn, error)
	io.Closer
}

type manager struct {
	log            *log.Logger
	tlsConfig      *tls.Config
	lock           *sync.Mutex
	connectionPool map[string]*grpc.ClientConn
}

// NewGRPCManager
func NewManager(cert utils.Certificate) (Manager, error) {
	tlsConfig, err := utils.NewTLSConfigClient(cert)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &manager{
		log:            log.With(log.Any("main", "manager")),
		tlsConfig:      tlsConfig,
		lock:           &sync.Mutex{},
		connectionPool: map[string]*grpc.ClientConn{},
	}, nil
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
		grpc.WithTransportCredentials(credentials.NewTLS(g.tlsConfig)),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		g.log.Error("failed to create connection to server", log.Error(err), log.Any("address", address))
		return nil, errors.Trace(err)
	}

	g.connectionPool[address] = conn
	return conn, nil
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
