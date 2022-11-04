package reward

import (
	"context"
	"errors"
	"fmt"
	mabi "github.com/redesblock/mop/core/contract/abi"
	"github.com/redesblock/mop/core/mctx"
	"github.com/redesblock/mop/core/storer/storage"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/redesblock/mop/core/chain/transaction"
)

var (
	rewardABI            = transaction.ParseABIUnchecked(mabi.RewardABIv0_1_0)
	errDecodeABI         = errors.New("could not decode abi data")
	ErrInsufficientFunds = errors.New("insufficient token balance")
	keyPrefix            = "reward-txs-"
)

type Service interface {
	TokenAddress(ctx context.Context) (common.Address, error)
	SystemTokenAddress(ctx context.Context) (common.Address, error)
	GetSystemReward(ctx context.Context, address common.Address) (*big.Int, error)
	GetCashedReward(ctx context.Context, address common.Address) (*big.Int, error)
	GetUnCashReward(ctx context.Context, address common.Address) (*big.Int, error)
	DoSystemReward(ctx context.Context, addresses []common.Address, values []*big.Int) (common.Hash, error)
	DoReward(ctx context.Context, addresses []common.Address, values []*big.Int) (common.Hash, error)
	Cash(ctx context.Context, address common.Address, value *big.Int) (common.Hash, error)
	Txs() ([]string, error)
}

type service struct {
	stateStore         storage.StateStorer
	transactionService transaction.Service
	address            common.Address
}

func New(stateStore storage.StateStorer, transactionService transaction.Service, address common.Address) Service {
	return &service{
		stateStore:         stateStore,
		transactionService: transactionService,
		address:            address,
	}
}

func (c *service) TokenAddress(ctx context.Context) (common.Address, error) {
	callData, err := rewardABI.Pack("token")
	if err != nil {
		return common.Address{}, err
	}

	output, err := c.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &c.address,
		Data: callData,
	})
	if err != nil {
		return common.Address{}, err
	}

	results, err := rewardABI.Unpack("token", output)
	if err != nil {
		return common.Address{}, err
	}

	if len(results) != 1 {
		return common.Address{}, errDecodeABI
	}

	erc20Address, ok := abi.ConvertType(results[0], new(common.Address)).(*common.Address)
	if !ok || erc20Address == nil {
		return common.Address{}, errDecodeABI
	}
	return *erc20Address, nil
}

func (c *service) SystemTokenAddress(ctx context.Context) (common.Address, error) {
	callData, err := rewardABI.Pack("systemToken")
	if err != nil {
		return common.Address{}, err
	}

	output, err := c.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &c.address,
		Data: callData,
	})
	if err != nil {
		return common.Address{}, err
	}

	results, err := rewardABI.Unpack("systemToken", output)
	if err != nil {
		return common.Address{}, err
	}

	if len(results) != 1 {
		return common.Address{}, errDecodeABI
	}

	erc20Address, ok := abi.ConvertType(results[0], new(common.Address)).(*common.Address)
	if !ok || erc20Address == nil {
		return common.Address{}, errDecodeABI
	}
	return *erc20Address, nil
}

func (s *service) GetSystemReward(ctx context.Context, address common.Address) (*big.Int, error) {
	callData, err := rewardABI.Pack("systemReward", address)
	if err != nil {
		return nil, err
	}

	output, err := s.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &s.address,
		Data: callData,
	})
	if err != nil {
		return nil, err
	}

	results, err := rewardABI.Unpack("systemReward", output)
	if err != nil {
		return nil, err
	}

	if len(results) != 1 {
		return nil, errDecodeABI
	}

	balance, ok := abi.ConvertType(results[0], new(big.Int)).(*big.Int)
	if !ok || balance == nil {
		return nil, errDecodeABI
	}
	return balance, nil
}

func (s *service) GetCashedReward(ctx context.Context, address common.Address) (*big.Int, error) {
	callData, err := rewardABI.Pack("cashedReward", address)
	if err != nil {
		return nil, err
	}

	output, err := s.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &s.address,
		Data: callData,
	})
	if err != nil {
		return nil, err
	}

	results, err := rewardABI.Unpack("cashedReward", output)
	if err != nil {
		return nil, err
	}

	if len(results) != 1 {
		return nil, errDecodeABI
	}

	balance, ok := abi.ConvertType(results[0], new(big.Int)).(*big.Int)
	if !ok || balance == nil {
		return nil, errDecodeABI
	}
	return balance, nil
}

func (s *service) GetUnCashReward(ctx context.Context, address common.Address) (*big.Int, error) {
	callData, err := rewardABI.Pack("unCashReward", address)
	if err != nil {
		return nil, err
	}

	output, err := s.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &s.address,
		Data: callData,
	})
	if err != nil {
		return nil, err
	}

	results, err := rewardABI.Unpack("unCashReward", output)
	if err != nil {
		return nil, err
	}

	if len(results) != 1 {
		return nil, errDecodeABI
	}

	balance, ok := abi.ConvertType(results[0], new(big.Int)).(*big.Int)
	if !ok || balance == nil {
		return nil, errDecodeABI
	}
	return balance, nil
}

func (s *service) Txs() ([]string, error) {
	var txs []string
	if err := s.stateStore.Iterate(keyPrefix, func(k, v []byte) (bool, error) {
		if !strings.HasPrefix(string(k), keyPrefix) {
			return true, nil
		}

		tx := strings.TrimPrefix(string(k), keyPrefix)
		txs = append(txs, tx)
		return false, nil
	}); err != nil {
		return nil, err
	}

	return txs, nil
}

func (s *service) storeTx(ctx context.Context, txHash common.Hash, wait bool) error {
	if wait {
		receipt, err := s.transactionService.WaitForReceipt(ctx, txHash)
		if err != nil {
			return err
		}

		s.stateStore.Put(keyPrefix+txHash.String(), receipt)

		if receipt.Status == 0 {
			return transaction.ErrTransactionReverted
		}
	} else {
		s.stateStore.Put(keyPrefix+txHash.String(), txHash)
	}
	return nil
}

func (c *service) Cash(ctx context.Context, address common.Address, value *big.Int) (common.Hash, error) {
	balance, err := c.GetUnCashReward(ctx, address)
	if err != nil {
		return common.Hash{}, err
	}

	// check we can afford this so we don't waste gas and don't risk bouncing cheques
	if balance.Cmp(value) < 0 {
		return common.Hash{}, ErrInsufficientFunds
	}

	callData, err := rewardABI.Pack("cash", value)
	if err != nil {
		return common.Hash{}, err
	}

	request := &transaction.TxRequest{
		To:          &c.address,
		Data:        callData,
		GasPrice:    mctx.GetGasPrice(ctx),
		GasLimit:    90000,
		Value:       big.NewInt(0),
		Description: "withdraw reward",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	if err := c.storeTx(ctx, txHash, false); err != nil {
		return common.Hash{}, err
	}

	return txHash, nil
}

func (c *service) DoSystemReward(ctx context.Context, addresses []common.Address, values []*big.Int) (common.Hash, error) {
	if len(addresses) != len(values) {
		return common.Hash{}, fmt.Errorf("mismatch num")
	}

	callData, err := rewardABI.Pack("doSystemToken", addresses, values)
	if err != nil {
		return common.Hash{}, err
	}

	request := &transaction.TxRequest{
		To:          &c.address,
		Data:        callData,
		GasPrice:    mctx.GetGasPrice(ctx),
		GasLimit:    100000,
		Value:       big.NewInt(0),
		Description: "system reward",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	if err := c.storeTx(ctx, txHash, false); err != nil {
		return common.Hash{}, err
	}

	return txHash, nil
}

func (c *service) DoReward(ctx context.Context, addresses []common.Address, values []*big.Int) (common.Hash, error) {
	if len(addresses) != len(values) {
		return common.Hash{}, fmt.Errorf("mismatch num")
	}

	callData, err := rewardABI.Pack("doToken", addresses, values)
	if err != nil {
		return common.Hash{}, err
	}

	request := &transaction.TxRequest{
		To:          &c.address,
		Data:        callData,
		GasPrice:    mctx.GetGasPrice(ctx),
		GasLimit:    900000,
		Value:       big.NewInt(0),
		Description: "system reward",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	if err := c.storeTx(ctx, txHash, false); err != nil {
		return common.Hash{}, err
	}

	return txHash, nil
}
