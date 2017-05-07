package servicer

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gomatic/go-vbuild"
	"github.com/gomatic/usage"
	"github.com/urfave/cli"
)

// Config type.
type Config func(*cli.App) error

// ConfigFunc type.
type ConfigFunc func(*cli.App) cli.ActionFunc

// ErrorFunc type.
func ErrorFunc(err error) cli.ActionFunc {
	return func(ctx *cli.Context) error { return err }
}

// Main entry-point for servicers.
func Main(configure ConfigFunc) {
	settings.Version = build.Version.String()
	app := cli.NewApp()
	app.ArgsUsage = ""
	app.Version = settings.Version
	app.EnableBashCompletion = true
	app.OnUsageError = usage.Error

	readTimtout, writeTimeout := 0, 0

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "api-port, api",
			Usage:       "API Port.",
			EnvVar:      "API_PORT",
			Value:       3000,
			Destination: &settings.Api.Port,
		},
		cli.IntFlag{
			Name:        "rpc-port, rpc",
			Usage:       "RPC Port. Defaults to main-port + 1.",
			EnvVar:      "RPC_PORT",
			Value:       -1,
			Destination: &settings.Rpc.Port,
		},
		cli.StringFlag{
			Name:        "bind-addr, bind",
			Usage:       "Bind Address.",
			EnvVar:      "APP_ADDR",
			Value:       "127.0.0.1",
			Destination: &settings.Bind,
		},
		cli.StringFlag{
			Name:        "api-addr",
			Usage:       "API Address. Defaults so bind-addr.",
			EnvVar:      "API_ADDR",
			Value:       "",
			Destination: &settings.Api.Addr,
		},
		cli.StringFlag{
			Name:        "rpc-addr",
			Usage:       "RPC Address. Defaults so bind-addr.",
			EnvVar:      "RPC_ADDR",
			Value:       "",
			Destination: &settings.Rpc.Addr,
		},
		cli.StringFlag{
			Name:        "namespace, ns",
			Usage:       "Service namespace.",
			EnvVar:      "SVC_NAMESPACE",
			Value:       "dev",
			Destination: &settings.Dns.Namespace,
		},
		cli.StringFlag{
			Name:        "name",
			Usage:       "Server name.",
			Value:       path.Base(os.Args[0]),
			Destination: &settings.Name,
		},
		cli.StringFlag{
			Name:        "domain",
			Usage:       "Service domain.",
			EnvVar:      "SVC_DOMAIN",
			Value:       "svc.cluster.local",
			Destination: &settings.Dns.Domain,
		},
		cli.IntFlag{
			Name:        "readTimtout-timeout",
			Usage:       "Read timeout in seconds.",
			EnvVar:      "READ_TIMEOUT",
			Value:       5,
			Destination: &readTimtout,
		},
		cli.IntFlag{
			Name:        "writeTimeout-timeout",
			Usage:       "Write timeout in seconds.",
			EnvVar:      "WRITE_TIMEOUT",
			Value:       10,
			Destination: &writeTimeout,
		},
		cli.BoolFlag{
			Name:        "verbose, V",
			Usage:       "Verbose output.",
			EnvVar:      "VERBOSE",
			Destination: &settings.Output.Verbose,
		},
		cli.BoolFlag{
			Name:        "mock, M",
			Usage:       "Return mock JSON.",
			Destination: &settings.Output.Mocking,
		},
		cli.BoolFlag{
			Name:        "debug, D",
			Usage:       "Enable debugging.",
			EnvVar:      "DEBUG",
			Destination: &settings.Output.Debugging,
		},
	}

	app.Before = func(ctx *cli.Context) error {

		func() {
			// cat /proc/self/cgroup | awk -F'/' '{print $3}'
			f, err := os.Open("/proc/self/cgroup")
			if err != nil {
				return
			}
			defer f.Close()
			r := bufio.NewReader(f)
			first, err := r.ReadString('\n')
			if err != nil {
				return
			}
			if parts := strings.Split(strings.TrimSpace(first), "/"); len(parts) < 3 || parts[1] != "docker" {
				log.Printf("Not a properly formatted /proc/self/cgroup file: %+v", parts)
			} else {
				settings.Container = parts[2]
			}
		}()

		if settings.Container != "" {
			settings.Bind = "0.0.0.0"
			if settings.Api.Port <= 0 {
				settings.Api.Port = 3000
			}
		}

		if settings.Api.Addr == "" {
			settings.Api.Addr = settings.Bind
		}
		if settings.Rpc.Addr == "" {
			settings.Rpc.Addr = settings.Bind
		}
		if settings.Api.Port <= 0 {
			rand.Seed(time.Now().UTC().UnixNano())
			settings.Api.Port = rand.Intn(16380) + 49152
		}
		if settings.Rpc.Port <= 0 {
			settings.Rpc.Port = settings.Api.Port + 1
		}

		maxTimeout := 300

		if readTimtout <= 0 {
			readTimtout = 5
		} else if readTimtout > maxTimeout {
			readTimtout = maxTimeout
		}
		settings.Timeout.Read = time.Duration(readTimtout) * time.Second

		if writeTimeout <= 0 {
			writeTimeout = 10
		} else if writeTimeout > maxTimeout {
			writeTimeout = maxTimeout
		}
		settings.Timeout.Write = time.Duration(writeTimeout) * time.Second

		log.Printf("settings: %+v", settings)
		ctx.App.Metadata["settings"] = settings

		return nil
	}

	app.Action = usage.Trapper(debugger(configure(app)))

	app.Version = app.Version + "." + settings.Version
	app.Run(os.Args)
}
