package debugapi

import (
	"math/big"
	"net/http"

	"github.com/redesblock/hop/core/jsonhttp"
)

var (
	errChequebookBalance = "cannot get chequebook balance"
)

type chequebookBalanceResponse struct {
	Balance *big.Int `json:"balance"`
}

type chequebookAddressResponse struct {
	Address string `json:"chequebookaddress"`
}

func (s *server) chequebookBalanceHandler(w http.ResponseWriter, r *http.Request) {
	balance, err := s.Chequebook.Balance(r.Context())
	if err != nil {
		jsonhttp.InternalServerError(w, errChequebookBalance)
		s.Logger.Debugf("debug api: chequebook balance: %v", err)
		s.Logger.Error("debug api: cannot get chequebook balance")
		return
	}

	jsonhttp.OK(w, chequebookBalanceResponse{Balance: balance})
}

func (s *server) chequebookAddressHandler(w http.ResponseWriter, r *http.Request) {
	address := s.Chequebook.Address()
	jsonhttp.OK(w, chequebookAddressResponse{Address: address.String()})
}
