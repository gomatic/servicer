package server

import (
	"net"
	"net/http"

	"github.com/gomatic/servicer"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//
type MainFunc func(context.Context, servicer.Settings, *grpc.Server) (*runtime.ServeMux, error)

//
type Runner func(MainFunc) cli.ActionFunc

//
func Main(main MainFunc, name, usage string) {
	servicer.Main(run(main), name, usage)
}

//
func run(main MainFunc) cli.ActionFunc {
	return func(app *cli.Context) error {

		settings := app.App.Metadata["settings"].(servicer.Settings)

		rpcHost := settings.Rpc.String()
		rpcListener, err := net.Listen("tcp", rpcHost)
		if err != nil {
			return err
		}

		rpcServer := grpc.NewServer()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mux, err := main(ctx, settings, rpcServer)
		if err != nil {
			return err
		}

		go rpcServer.Serve(rpcListener)

		s := &http.Server{
			Addr:           settings.Api.String(),
			Handler:        mux,
			ReadTimeout:    settings.Timeout.Read,
			WriteTimeout:   settings.Timeout.Write,
			MaxHeaderBytes: 1 << 20,
		}

		return s.ListenAndServe()
	}
}
