package servicer

import (
	"fmt"
	"time"
)

//
const MAJOR = "1"
var VERSION = "0"

//
type host struct {
	Port int
	Addr string
}

//
func (h host) String() string {
	return fmt.Sprintf("%s:%d", h.Addr, h.Port)
}

//
type Settings struct {
	Bind string
	Api  host
	Rpc  host

	Dns struct {
		Namespace, Domain string
	}

	Container string
	Version   string

	Timeout struct {
		Read, Write time.Duration
	}

	Output struct {
		Mocking   bool
		Debugging bool
		Verbose   bool
	}
}

//
var settings Settings
