package usage

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/urfave/cli"
)

//
func Error(ctx *cli.Context, err error, isSubcommand bool) error {
	if isSubcommand {
		cli.ShowSubcommandHelp(ctx)
	} else {
		cli.ShowAppHelp(ctx)
	}
	fmt.Fprintf(os.Stderr, "%v\n", err)
	return nil
}

//
func Trapper(action cli.ActionFunc) cli.ActionFunc {
	started := time.Now()
	return func(ctx *cli.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
				if os.Getenv("DEBUG") != "" {
					debug.PrintStack()
				}
			}

			if err != nil {
				Error(ctx, cli.NewExitError(err.Error(), 1), true)
			}

			elapsed := time.Now().Sub(started)
			log.Printf("elapsed: %v", elapsed)
		}()

		err = action(ctx)
		return
	}
}
