package debugapi

import (
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/bigint"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/sctx"
)

const (
	errETHBalance                 = "cannot get chain balance"
	errTotalPledgedBalanceBalance = "cannot get total pledged balance"
	errPledgedBalanceBalance      = "cannot get pledged balance"
	errTotalSlashedBalanceBalance = "cannot get total slashed balance"
	errSlashedBalanceBalance      = "cannot get slashed balance"
)

type BeeNodeMode uint

const (
	LightMode BeeNodeMode = iota
	FullMode
	DevMode
	UltraLightMode
)

type nodeResponse struct {
	BeeMode           string `json:"beeMode"`
	GatewayMode       bool   `json:"gatewayMode"`
	ChequebookEnabled bool   `json:"chequebookEnabled"`
	SwapEnabled       bool   `json:"swapEnabled"`
}

func (b BeeNodeMode) String() string {
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

// nodeGetHandler gives back information about the Bee node configuration.
func (s *Service) nodeGetHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, nodeResponse{})
}

type nodeBalanceResponse struct {
	Balance *bigint.BigInt `json:"balance"`
}

func (s *Service) nodeBalanceHandler(w http.ResponseWriter, r *http.Request) {
	balance, err := s.backend.BalanceAt(r.Context(), s.ethereumAddress, nil)
	if err != nil {
		jsonhttp.InternalServerError(w, errETHBalance)
		s.logger.Error(r.URL.Path, err)
		return
	}
	jsonhttp.OK(w, nodeBalanceResponse{Balance: bigint.Wrap(balance)})
}

type nodeStorageResponse struct {
	SystemBalance *bigint.BigInt `json:"systemBalance"`
	CashBalance   *bigint.BigInt `json:"cashBalance"`
	UncashBalance *bigint.BigInt `json:"unCashBalance"`
}

func (s *Service) nodeStorageHandler(w http.ResponseWriter, r *http.Request) {
	balance := big.NewInt(0)
	jsonhttp.OK(w, nodeStorageResponse{SystemBalance: bigint.Wrap(balance), CashBalance: bigint.Wrap(balance), UncashBalance: bigint.Wrap(balance)})
}

type nodeStorageTransactionResponse struct {
	Txs []string `json:"txs"`
}

func (s *Service) nodeStorageTransactionHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, nodeStorageTransactionResponse{
		Txs: []string{"0x03ef42927e896883ec2666cb1f0b6d758136c7f08eebd8e620a44a820ea86d2fda"},
	})
}

type nodePledgeResponse struct {
	TotalPledgedBalance *bigint.BigInt `json:"totalPledgedBalance"`
	PledgedBalance      *bigint.BigInt `json:"pledgedBalance"`
	TotalSlashedBalance *bigint.BigInt `json:"totalSlashedBalance"`
	SlashedBalance      *bigint.BigInt `json:"slashedBalance"`
}

func (s *Service) nodePledgeHandler(w http.ResponseWriter, r *http.Request) {
	totalPledgedBalance, err := s.pledgeContract.GetTotalShare(r.Context())
	if err != nil {
		jsonhttp.InternalServerError(w, errTotalPledgedBalanceBalance)
		s.logger.Error(r.URL.Path, err)
		return
	}

	pledgedBalance, err := s.pledgeContract.GetShare(r.Context(), s.ethereumAddress)
	if err != nil {
		jsonhttp.InternalServerError(w, errPledgedBalanceBalance)
		s.logger.Error(r.URL.Path, err)
		return
	}

	totalSlashedBalance, err := s.pledgeContract.GetTotalSlash(r.Context())
	if err != nil {
		jsonhttp.InternalServerError(w, errTotalSlashedBalanceBalance)
		s.logger.Error(r.URL.Path, err)
		return
	}

	slashedBalance, err := s.pledgeContract.GetSlash(r.Context(), s.ethereumAddress)
	if err != nil {
		jsonhttp.InternalServerError(w, errSlashedBalanceBalance)
		s.logger.Error(r.URL.Path, err)
		return
	}

	jsonhttp.OK(w, nodePledgeResponse{
		TotalPledgedBalance: bigint.Wrap(totalPledgedBalance),
		PledgedBalance:      bigint.Wrap(pledgedBalance),
		TotalSlashedBalance: bigint.Wrap(totalSlashedBalance),
		SlashedBalance:      bigint.Wrap(slashedBalance),
	})
}

type nodePledgeTransactionResponse struct {
	Txs []string `json:"txs"`
}

func (s *Service) nodePledgeTransactionHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.OK(w, nodePledgeTransactionResponse{
		Txs: []string{"0x26304c4f2924f1f50aec75d03f083d7d9caf9c49c77ceccb594cfd61df616128"},
	})
}

type pledgeTxResponse struct {
	TransactionHash common.Hash `json:"transactionHash"`
}

func (s *Service) nodePledgeStakeHandler(w http.ResponseWriter, r *http.Request) {
	amount, ok := big.NewInt(0).SetString(mux.Vars(r)["amount"], 10)
	if !ok {
		s.logger.Error("create batch: invalid amount")
		jsonhttp.BadRequest(w, "invalid postage amount")
		return

	}

	ctx := r.Context()
	if price, ok := r.Header[gasPriceHeader]; ok {
		p, ok := big.NewInt(0).SetString(price[0], 10)
		if !ok {
			s.logger.Error("debug api: withdraw: bad gas price")
			jsonhttp.BadRequest(w, errBadGasPrice)
			return
		}
		ctx = sctx.SetGasPrice(ctx, p)
	}

	txhash, err := s.pledgeContract.Stake(ctx, amount)
	if err != nil {
		s.logger.Error(r.URL.Path, err)
		jsonhttp.InternalServerError(w, "cannot stake")
		return
	}

	jsonhttp.OK(w, &pledgeTxResponse{
		TransactionHash: txhash,
	})
}

func (s *Service) nodePledgeUnStakeHandler(w http.ResponseWriter, r *http.Request) {
	amount, ok := big.NewInt(0).SetString(mux.Vars(r)["amount"], 10)
	if !ok {
		s.logger.Error("create batch: invalid amount")
		jsonhttp.BadRequest(w, "invalid postage amount")
		return

	}

	ctx := r.Context()
	if price, ok := r.Header[gasPriceHeader]; ok {
		p, ok := big.NewInt(0).SetString(price[0], 10)
		if !ok {
			s.logger.Error("debug api: withdraw: bad gas price")
			jsonhttp.BadRequest(w, errBadGasPrice)
			return
		}
		ctx = sctx.SetGasPrice(ctx, p)
	}

	txhash, err := s.pledgeContract.UnStake(ctx, amount)
	if err != nil {
		s.logger.Error(r.URL.Path, err)
		jsonhttp.InternalServerError(w, "cannot unstake")
		return
	}

	jsonhttp.OK(w, &pledgeTxResponse{
		TransactionHash: txhash,
	})
}
