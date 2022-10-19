package client

import (
	"github.com/redesblock/mop/core/resolver"
)

// Interface is a resolver client that can connect/disconnect to an external
// Name Resolution Service via an endpoint.
type Interface interface {
	resolver.Interface
	Endpoint() string
	IsConnected() bool
}
