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
			return err
		}

		var resolver resolve.Resolver
		switch ctx.RunMode() {
		case context.RunModeKube:
			resolver, err = resolve.NewKubeResolver(ctx)
			if err != nil {
				return err
			}
		case context.RunModeNative:
			resolver, err = resolve.NewNativeResolver(ctx)
			if err != nil {
				return err
			}
		default:
			return errors.Errorf("Run mode (%s) is not supported.", ctx.RunMode())
		}

		api, err := function.NewAPI(cfg, ctx, resolver)
		if err != nil {
			return err
		}
		defer api.Close()
		ctx.Wait()
		return nil
	})
}
