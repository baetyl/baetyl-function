package resolve

import (
	"fmt"

	"github.com/baetyl/baetyl-go/v2/context"
)

type KubeResolver struct {
	ctx context.Context
}

func NewKubeResolver(ctx context.Context) (Resolver, error) {
	return &KubeResolver{ctx: ctx}, nil
}

func (k *KubeResolver) Resolve(service string) (address string, err error) {
	return fmt.Sprintf("%s.%s", service, k.ctx.EdgeNamespace()), nil
}

func (k *KubeResolver) Close() error {
	return nil
}
