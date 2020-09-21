package resolve

import (
	"io"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
)

// Factories of Resolver
var Factories = map[string]func(ctx context.Context) (Resolver, error){}

type Resolver interface {
	Resolve(service string) (address string, err error)
	io.Closer
}

// New Resolver by params
func New(mode string, ctx context.Context) (Resolver, error) {
	if f, ok := Factories[mode]; ok {
		return f(ctx)
	}
	return nil, errors.Errorf("factory didn't found resolver according to the mode (%s)", mode)
}
