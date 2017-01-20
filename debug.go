package servicer

import (
	"expvar"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"github.com/urfave/cli"
)

//
func debugger(action cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		settings := ctx.App.Metadata["settings"].(Settings)

		if !settings.Output.Debugging && settings.Container == "" {
			return action(ctx)
		}

		env := make(map[string]string)
		for _, item := range os.Environ() {
			splits := strings.Split(item, "=")
			env[splits[0]] = splits[1]
		}

		port := strconv.Itoa(settings.Api.Port - 1)
		expvar.Publish("env", expvar.Func(func() interface{} { return env }))
		expvar.Publish("settings", expvar.Func(func() interface{} { return settings }))
		go func() {
			srv := &http.Server{
				Addr:         settings.Bind + ":" + port,
				ReadTimeout:  settings.Timeout.Read,
				WriteTimeout: settings.Timeout.Write,
			}
			log.Println("debugging on: " + srv.Addr)
			log.Println(srv.ListenAndServe())
		}()

		return action(ctx)
	}
}
