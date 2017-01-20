package client

import (
	"github.com/gomatic/servicer"
	"github.com/urfave/cli"
)

//
type MainFunc func(servicer.Settings) error

//
type Runner func(MainFunc) cli.ActionFunc

//
func Main(main MainFunc) cli.ActionFunc {
	return func(app *cli.Context) error {
		return main(app.App.Metadata["settings"].(servicer.Settings))
	}
}
