package api_test

import (
	"testing"

	"github.com/redesblock/mop/core/api"
)

func TestMopNodeMode_String(t *testing.T) {
	const nonExistingMode api.MopNodeMode = 4

	mapping := map[string]string{
		api.LightMode.String():   "light",
		api.FullMode.String():    "full",
		api.DevMode.String():     "dev",
		nonExistingMode.String(): "unknown",
	}

	for have, want := range mapping {
		if have != want {
			t.Fatalf("unexpected mop node mode: have %q; want %q", have, want)
		}
	}
}
