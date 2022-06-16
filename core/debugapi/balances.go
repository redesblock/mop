package debugapi

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/accounting"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/swarm"
)

var (
	errCantBalances  = "Cannot get balances"
	errCantBalance   = "Cannot get balance"
	errNoBalance     = "No balance for peer"
	errInvaliAddress = "Invalid address"
)

type balanceResponse struct {
	Peer    string `json:"peer"`
	Balance int64  `json:"balance"`
}

type balancesResponse struct {
	Balances []balanceResponse `json:"balances"`
}

func (s *server) balancesHandler(w http.ResponseWriter, r *http.Request) {
	balances, err := s.Accounting.Balances()
	if err != nil {
		jsonhttp.InternalServerError(w, errCantBalances)
		s.Logger.Debugf("debug api: balances: %v", err)
		s.Logger.Error("debug api: can not get balances")
		return
	}

	balResponses := make([]balanceResponse, len(balances))
	i := 0
	for k := range balances {
		balResponses[i] = balanceResponse{
			Peer:    k,
			Balance: balances[k],
		}
		i++
	}

	jsonhttp.OK(w, balancesResponse{Balances: balResponses})
}

func (s *server) peerBalanceHandler(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["peer"]
	peer, err := swarm.ParseHexAddress(addr)
	if err != nil {
		s.Logger.Debugf("debug api: balances peer: invalid peer address %s: %v", addr, err)
		s.Logger.Errorf("debug api: balances peer: invalid peer address %s", addr)
		jsonhttp.NotFound(w, errInvaliAddress)
		return
	}

	balance, err := s.Accounting.Balance(peer)
	if err != nil {
		if errors.Is(err, accounting.ErrPeerNoBalance) {
			jsonhttp.NotFound(w, errNoBalance)
			return
		}
		s.Logger.Debugf("debug api: balances peer: get peer %s balance: %v", peer.String(), err)
		s.Logger.Errorf("debug api: balances peer: can't get peer %s balance", peer.String())
		jsonhttp.InternalServerError(w, errCantBalance)
		return
	}

	jsonhttp.OK(w, balanceResponse{
		Peer:    peer.String(),
		Balance: balance,
	})
}
