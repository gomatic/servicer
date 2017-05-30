package servicer

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pborman/uuid"
	"github.com/urfave/cli"
)

// TODO unify with gateway/run.go
var (
	signingSecret = []byte("secret")
	signingMethod = jwt.SigningMethodHS256
	iss           = "gateway"
)

func init() {
	if s, exists := os.LookupEnv("SECRET"); exists {
		signingSecret = []byte(s)
	}
}

//
func createToken(req *http.Request) (string, error) {
	now := time.Now().UTC()
	duration := time.Duration(5)
	q := req.URL.Query()
	if d := q.Get("d"); d != "" {
		if dur, err := strconv.Atoi(d); err == nil && dur < 60 && dur > 1 {
			duration = time.Duration(dur)
		}
	}

	sub := uuid.NewRandom().String()
	iat := now.Unix()

	// Create a secure, opaque aud so as not to expose internal services.

	h := hmac.New(md5.New, []byte(signingSecret))
	h.Write([]byte(sub))
	h.Write([]byte(iss))
	h.Write([]byte(fmt.Sprintf("%d", iat)))
	aud := hex.EncodeToString(h.Sum(nil))

	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims{
		"sub": sub,
		"aud": aud,
		"iss": iss,
		"iat": iat,
		"nbf": now.Add(-duration * time.Minute).Unix(),
		"exp": now.Add(duration * time.Minute).Unix(),
		"jti": uuid.NewRandom().String(),
	})

	// Sign and get the complete encoded token as a string using the signingSecret
	return token.SignedString(signingSecret)
}

//
func tokenizer(message string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		tokenString, err := createToken(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Signing failed")
			log.Printf("Token request: %+v\n%+v", req, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s%s", message, tokenString)
		return
	}

}

//
func debugger(action cli.ActionFunc) cli.ActionFunc {
	if action == nil {
		action = func(ctx *cli.Context) error {
			log.Println("WARNING: nil debugger ActionFunc")
			return nil
		}
	}
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
			mux := http.DefaultServeMux

			mux.HandleFunc("/token", tokenizer(""))
			mux.HandleFunc("/header", tokenizer("Authorization: Bearer "))

			srv := &http.Server{
				Addr:           settings.Bind + ":" + port,
				Handler:        mux,
				ReadTimeout:    settings.Timeout.Read,
				WriteTimeout:   settings.Timeout.Write,
				MaxHeaderBytes: 1 << 20,
			}

			log.Println("debugging on: " + srv.Addr)
			log.Println(srv.ListenAndServe())
		}()

		return action(ctx)
	}
}
