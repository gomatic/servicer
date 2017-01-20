package gateway

import (
	"github.com/gomatic/servicer"
	"github.com/urfave/cli"
)

//
type MainFunc func(servicer.Settings) error

//
type Runner func(MainFunc) cli.ActionFunc

//
func Main(main MainFunc, name, usage string) {
	servicer.Main(func(ctx *cli.Context) error {
		settings := ctx.App.Metadata["settings"].(servicer.Settings)
		return main(settings)
	}, name, usage)
}
