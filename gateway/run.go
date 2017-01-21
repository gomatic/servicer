package gateway

import (
	"log"

	"github.com/gomatic/servicer"
	"github.com/urfave/cli"
)

//
type MainFunc func(servicer.Settings) error

//
type Runner func(MainFunc) cli.ActionFunc

//
func Main(main MainFunc, config servicer.Config) {
	wrap := func(app *cli.App) cli.ActionFunc {
		if config != nil {
			if err := config(app); err != nil {
				log.Println(err)
				return servicer.ErrorFunc(err)
			}
		}
		return func(ctx *cli.Context) error {
			settings := ctx.App.Metadata["settings"].(servicer.Settings)
			return main(settings)
		}
	}
	servicer.Main(wrap)
}
