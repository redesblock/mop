package client

import (
	"github.com/redesblock/hop/core/resolver"
)

// Interface is a resolver client that can connect/disconnect to an external
// Name Resolution Service via an edpoint.
type Interface interface {
	resolver.Interface
	Connect(endpoint string) error
	IsConnected() bool
}
