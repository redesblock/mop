package debugapi

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/redesblock/hop/core/bigint"
	"github.com/redesblock/hop/core/jsonhttp"
	"github.com/redesblock/hop/core/sctx"
)

const (
	errETHBalance                 = "cannot get bnb balance"
	errHOPBalance                 = "cannot get hop balance"
	errMOPBalance                 = "cannot get mop balance"
	errTotalPledgedBalanceBalance = "cannot get total pledged balance"
	errPledgedBalanceBalance      = "cannot get pledged balance"
	errTotalSlashedBalanceBalance = "cannot get total slashed balance"
	errSlashedBalanceBalance      = "cannot get slashed balance"
	errSystemRewardBalanceBalance = "cannot get system reward balance"
	errCashedBalanceBalance       = "cannot get cashed reward balance"
	errUnCashBalanceBalance       = "cannot get uncash reward balance"
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
	Balance    *bigint.BigInt `json:"balance"`
	HopBalance *bigint.BigInt `json:"hopBalance"`
	MopBalance *bigint.BigInt `json:"mopBalance"`
}

func (s *Service) nodeBalanceHandler(w http.ResponseWriter, r *http.Request) {
	balance, err := s.backend.BalanceAt(r.Context(), s.ethereumAddress, nil)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errETHBalance)
		return
	}

	hopBalance, err := s.pledgeContract.AvailableBalance(r.Context(), s.ethereumAddress)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errHOPBalance)
		return
	}

	mopBalance, err := s.postageContract.AvailableBalance(r.Context(), s.ethereumAddress)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errMOPBalance)
		return
	}

	jsonhttp.OK(w, nodeBalanceResponse{Balance: bigint.Wrap(balance), HopBalance: bigint.Wrap(hopBalance), MopBalance: bigint.Wrap(mopBalance)})
}

type nodeRewardBalanceResponse struct {
	SystemBalance *bigint.BigInt `json:"systemBalance"`
	CashBalance   *bigint.BigInt `json:"cashBalance"`
	UncashBalance *bigint.BigInt `json:"unCashBalance"`
}

func (s *Service) nodeRewardBalanceHandler(w http.ResponseWriter, r *http.Request) {
	systemReward, err := s.rewardContract.GetSystemReward(r.Context(), s.ethereumAddress)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errSystemRewardBalanceBalance)
		return
	}

	cashedBalance, err := s.rewardContract.GetCashedReward(r.Context(), s.ethereumAddress)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errCashedBalanceBalance)
		return
	}

	uncashBalance, err := s.rewardContract.GetUnCashReward(r.Context(), s.ethereumAddress)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errUnCashBalanceBalance)
		return
	}

	jsonhttp.OK(w, nodeRewardBalanceResponse{SystemBalance: bigint.Wrap(systemReward), CashBalance: bigint.Wrap(cashedBalance), UncashBalance: bigint.Wrap(uncashBalance)})
}

type nodeRewardTransactionResponse struct {
	Txs []string `json:"txs"`
}

func (s *Service) nodeRewardTransactionHandler(w http.ResponseWriter, r *http.Request) {
	txs, err := s.rewardContract.Txs()
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, err)
		return
	}

	jsonhttp.OK(w, nodeRewardTransactionResponse{
		Txs: txs,
	})
}

type rewardTxResponse struct {
	TransactionHash common.Hash `json:"transactionHash"`
}

func (s *Service) nodeRewardCashHandler(w http.ResponseWriter, r *http.Request) {
	amount, ok := big.NewInt(0).SetString(mux.Vars(r)["amount"], 10)
	if !ok {
		s.logger.Error(r.URL.Path, "invalid amount")
		jsonhttp.BadRequest(w, "invalid amount")
		return
	}

	ctx := r.Context()
	if price, ok := r.Header[gasPriceHeader]; ok {
		p, ok := big.NewInt(0).SetString(price[0], 10)
		if !ok {
			s.logger.Error(r.URL.Path, "bad gas price")
			jsonhttp.BadRequest(w, errBadGasPrice)
			return
		}
		ctx = sctx.SetGasPrice(ctx, p)
	}

	txHash, err := s.rewardContract.Cash(ctx, amount)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, fmt.Sprintf("failed cash %v", err))
		return
	}

	jsonhttp.OK(w, &rewardTxResponse{
		TransactionHash: txHash,
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
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errTotalPledgedBalanceBalance)
		return
	}

	pledgedBalance, err := s.pledgeContract.GetShare(r.Context(), s.ethereumAddress)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errPledgedBalanceBalance)
		return
	}

	totalSlashedBalance, err := s.pledgeContract.GetTotalSlash(r.Context())
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errTotalSlashedBalanceBalance)
		return
	}

	slashedBalance, err := s.pledgeContract.GetSlash(r.Context(), s.ethereumAddress)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, errSlashedBalanceBalance)
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
	txs, err := s.pledgeContract.Txs()
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, err)
		return
	}

	jsonhttp.OK(w, nodePledgeTransactionResponse{
		Txs: txs,
	})
}

type pledgeTxResponse struct {
	TransactionHash common.Hash `json:"transactionHash"`
}

func (s *Service) nodePledgeStakeHandler(w http.ResponseWriter, r *http.Request) {
	amount, ok := big.NewInt(0).SetString(mux.Vars(r)["amount"], 10)
	if !ok {
		s.logger.Error(r.URL.Path, "invalid amount")
		jsonhttp.BadRequest(w, "invalid amount")
		return
	}

	ctx := r.Context()
	if price, ok := r.Header[gasPriceHeader]; ok {
		p, ok := big.NewInt(0).SetString(price[0], 10)
		if !ok {
			s.logger.Error(r.URL.Path, "bad gas price")
			jsonhttp.BadRequest(w, errBadGasPrice)
			return
		}
		ctx = sctx.SetGasPrice(ctx, p)
	}

	txHash, err := s.pledgeContract.Stake(ctx, amount)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, fmt.Sprintf("failed stake %v", err))
		return
	}

	jsonhttp.OK(w, &pledgeTxResponse{
		TransactionHash: txHash,
	})
}

func (s *Service) nodePledgeUnStakeHandler(w http.ResponseWriter, r *http.Request) {
	amount, ok := big.NewInt(0).SetString(mux.Vars(r)["amount"], 10)
	if !ok {
		s.logger.Error(r.URL.Path, "invalid amount")
		jsonhttp.BadRequest(w, "invalid amount")
		return
	}

	ctx := r.Context()
	if price, ok := r.Header[gasPriceHeader]; ok {
		p, ok := big.NewInt(0).SetString(price[0], 10)
		if !ok {
			s.logger.Error(r.URL.Path, "bad gas price")
			jsonhttp.BadRequest(w, errBadGasPrice)
			return
		}
		ctx = sctx.SetGasPrice(ctx, p)
	}

	txHash, err := s.pledgeContract.UnStake(ctx, amount)
	if err != nil {
		s.logger.Error(r.URL.Path, " error ", err)
		jsonhttp.InternalServerError(w, fmt.Sprintf("failed unstake %v", err))
		return
	}

	jsonhttp.OK(w, &pledgeTxResponse{
		TransactionHash: txHash,
	})
}
