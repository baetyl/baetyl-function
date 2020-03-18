package main

import (
	"github.com/baetyl/baetyl-go/context"
)

func main() {
	context.Run(func(ctx context.Context) error {
		var cfg Config
		err := ctx.LoadCustomConfig(&cfg)
		if err != nil {
			return err
		}

		api, err := NewAPI(cfg)
		if err != nil {
			return err
		}
		defer api.Close()
		ctx.Wait()
		return nil
	})
}
