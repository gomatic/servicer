package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gomatic/go-vbuild"
	"github.com/gomatic/servicer"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

var version = build.Version.String()

// MainFunc type.
type MainFunc func(context.Context, servicer.Settings, *grpc.Server) (*runtime.ServeMux, error)

// Runner type.
type Runner func(MainFunc) cli.ActionFunc

// Main client entry-point.
func Main(main MainFunc, name, usage string, api []byte) {
	config := func(app *cli.App) cli.ActionFunc {
		app.Name = name
		app.Usage = usage
		return run(main, api)
	}
	servicer.Main(config)
}

//
func serveAPI(api []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/model.json") {
			log.Printf("Not Found: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Write(api)
	}
}

//
func ok(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", version)
	return
}

//
func run(main MainFunc, api []byte) cli.ActionFunc {
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

		rmux, err := main(ctx, settings, rpcServer)
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/ok", ok)
		mux.HandleFunc("/health", ok)
		mux.HandleFunc("/health/", ok)
		mux.HandleFunc("/api/", serveAPI(api))
		mux.Handle("/", rmux)

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
