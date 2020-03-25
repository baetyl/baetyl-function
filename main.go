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

		api := NewAPI(cfg)
		defer api.Close()
		ctx.Wait()
		return nil
	})
}
