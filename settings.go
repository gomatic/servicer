package servicer

import (
	"fmt"
	"time"
)

//
type host struct {
	Port int
	Addr string
}

//
func (h host) String() string {
	return fmt.Sprintf("%s:%d", h.Addr, h.Port)
}

// Settings for the servicers.
type Settings struct {
	Bind string
	Api  host
	Rpc  host

	Dns struct {
		Namespace, Domain string
	}

	Name      string
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
