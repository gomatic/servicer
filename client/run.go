package client

import (
	"github.com/gomatic/servicer"
	"github.com/urfave/cli"
)

// MainFunc type.
type MainFunc func(servicer.Settings) error

// Runner type.
type Runner func(MainFunc) cli.ActionFunc

// Main client entry-point.
func Main(main MainFunc, name, usage string) {
	config := func(app *cli.App) cli.ActionFunc {
		app.Name = name
		app.Usage = usage
		return func(app *cli.Context) error {
			return main(app.App.Metadata["settings"].(servicer.Settings))
		}
	}
	servicer.Main(config)
}
