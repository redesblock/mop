package postage

import (
	"github.com/redesblock/hop/core/swarm"
)

func (st *StampIssuer) Inc(a swarm.Address) error {
	return st.inc(a)
}
