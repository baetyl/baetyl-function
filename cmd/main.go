package main

import (
	"github.com/baetyl/baetyl-function/v2/function"
	"github.com/baetyl/baetyl-go/v2/context"
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

		api, err := function.NewAPI(cfg, ctx)
		if err != nil {
			return err
		}
		defer api.Close()
		ctx.Wait()
		return nil
	})
}
