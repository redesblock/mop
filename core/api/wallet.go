package api

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/redesblock/mop/core/api/jsonhttp"
	"github.com/redesblock/mop/core/util/bigint"
)

type walletResponse struct {
	MOP             *bigint.BigInt `json:"mop"`             // the MOP balance of the wallet associated with the eth address of the node
	BNB             *bigint.BigInt `json:"bnb"`             // the bnb balance of the wallet associated with the eth address of the node
	ChainID         int64          `json:"chainID"`         // the id of the block chain
	ContractAddress common.Address `json:"contractAddress"` // the address of the chequebook contract
}

func (s *Service) walletHandler(w http.ResponseWriter, r *http.Request) {

	bnb, err := s.chainBackend.BalanceAt(r.Context(), s.bscAddress, nil)
	if err != nil {
		s.logger.Debug("wallet get: unable to acquire balance from the chain backend", "error", err)
		s.logger.Error(nil, "wallet get: unable to acquire balance from the chain backend")
		jsonhttp.InternalServerError(w, "unable to acquire balance from the chain backend")
		return
	}

	mop, err := s.erc20Service.BalanceOf(r.Context(), s.bscAddress)
	if err != nil {
		s.logger.Debug("wallet get: unable to acquire erc20 balance", "error", err)
		s.logger.Error(nil, "wallet get: unable to acquire erc20 balance")
		jsonhttp.InternalServerError(w, "unable to acquire erc20 balance")
		return
	}

	jsonhttp.OK(w, walletResponse{
		MOP:             bigint.Wrap(mop),
		BNB:             bigint.Wrap(bnb),
		ChainID:         s.chainID,
		ContractAddress: s.chequebook.Address(),
	})
}
