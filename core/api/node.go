package api

import (
	"net/http"

	"github.com/redesblock/mop/core/api/jsonhttp"
)

type MopNodeMode uint

const (
	LightMode MopNodeMode = iota
	FullMode
	DevMode
	UltraLightMode
)

type nodeResponse struct {
	MopMode           string `json:"mopMode"`
	GatewayMode       bool   `json:"gatewayMode"`
	ChequebookEnabled bool   `json:"chequebookEnabled"`
	SwapEnabled       bool   `json:"swapEnabled"`
}

func (b MopNodeMode) String() string {
	switch b {
	case LightMode:
		return "light"
	case FullMode:
		return "full"
	case DevMode:
		return "dev"
	case UltraLightMode:
		return "ultra-light"
	}
	return "unknown"
}

// nodeGetHandler gives back information about the Mop node configuration.
func (s *Service) nodeGetHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, nodeResponse{
		MopMode:           s.mopMode.String(),
		GatewayMode:       s.gatewayMode,
		ChequebookEnabled: s.chequebookEnabled,
		SwapEnabled:       s.swapEnabled,
	})
}
