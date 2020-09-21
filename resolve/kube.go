package resolve

import (
	"fmt"

	"github.com/baetyl/baetyl-go/v2/context"
)

func init() {
	Factories["kube"] = newKubeResolver
}

type kubeResolver struct {
	ctx context.Context
}

func newKubeResolver(ctx context.Context) (Resolver, error) {
	return &kubeResolver{ctx: ctx}, nil
}

func (k *kubeResolver) Resolve(service string) (address string, err error) {
	return fmt.Sprintf("%s.%s", service, k.ctx.EdgeNamespace()), nil
}

func (k *kubeResolver) Close() error {
	return nil
}
