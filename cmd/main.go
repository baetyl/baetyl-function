package main

import (
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"

	"github.com/baetyl/baetyl-function/v2/function"
	"github.com/baetyl/baetyl-function/v2/resolve"
)

func main() {
	context.Run(func(ctx context.Context) error {
		if err := ctx.CheckSystemCert(); err != nil {
			return err
		}

		var cfg function.Config
		err := ctx.LoadCustomConfig(&cfg)
		if err != nil {
			return errors.Trace(err)
		}

		resolver, err := resolve.New(context.RunMode(), ctx)
		if err != nil {
			return errors.Trace(err)
		}
		defer resolver.Close()

		api, err := function.NewAPI(cfg, ctx, resolver)
		if err != nil {
			return errors.Trace(err)
		}
		defer api.Close()
		ctx.Wait()
		return nil
	})
}
