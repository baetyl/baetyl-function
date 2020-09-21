package resolve

import (
	"fmt"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/native"
)

func init() {
	Factories["native"] = newNativeResolver
}

type nativeResolver struct {
	mapping *native.ServiceMapping
	log     *log.Logger
}

func newNativeResolver(_ context.Context) (Resolver, error) {
	logger := log.With(log.Any("resolve", "native"))
	mapping, err := native.NewServiceMapping()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = mapping.WatchFile(logger)
	if err != nil {
		return nil, err
	}

	return &nativeResolver{
		mapping: mapping,
		log:     log.With(log.Any("resolve", "native")),
	}, nil
}

func (n *nativeResolver) Resolve(service string) (address string, err error) {
	port, err := n.mapping.GetServiceNextPort(service)
	if err != nil {
		return "", errors.Trace(err)
	}

	return fmt.Sprintf("127.0.0.1:%d", port), nil
}

func (n *nativeResolver) Close() error {
	if n.mapping != nil {
		n.mapping.Close()
	}
	return nil
}
