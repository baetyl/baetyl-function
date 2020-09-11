package resolve

import "io"

type Resolver interface {
	Resolve(service string) (address string, err error)
	io.Closer
}
