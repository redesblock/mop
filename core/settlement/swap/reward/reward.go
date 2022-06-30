package reward

import (
	"context"
	"errors"
	"fmt"
	hopabi "github.com/redesblock/hop/contracts/abi"
	"github.com/redesblock/hop/core/sctx"
	"github.com/redesblock/hop/core/storage"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/redesblock/hop/core/transaction"
)

var (
	rewardABI            = transaction.ParseABIUnchecked(hopabi.RewardABI)
	errDecodeABI         = errors.New("could not decode abi data")
	ErrInsufficientFunds = errors.New("insufficient token balance")
	keyPrefix            = "reward-txs-"
)

type Service interface {
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
	backend            transaction.Backend
	transactionService transaction.Service
	address            common.Address
}

func New(stateStore storage.StateStorer, backend transaction.Backend, transactionService transaction.Service, address common.Address) Service {
	return &service{
		stateStore:         stateStore,
		backend:            backend,
		transactionService: transactionService,
		address:            address,
	}
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

func (s *service) storeTx(ctx context.Context, txHash common.Hash) error {
	receipt, err := s.transactionService.WaitForReceipt(ctx, txHash)
	if err != nil {
		return err
	}

	if s.stateStore != nil {
		s.stateStore.Put(keyPrefix+txHash.String(), receipt)
	}

	if receipt.Status == 0 {
		return transaction.ErrTransactionReverted
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
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    90000,
		Value:       big.NewInt(0),
		Description: "withdraw reward",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	if err := c.storeTx(ctx, txHash); err != nil {
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
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    90000,
		Value:       big.NewInt(0),
		Description: "system reward",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	if err := c.storeTx(ctx, txHash); err != nil {
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
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    90000,
		Value:       big.NewInt(0),
		Description: "system reward",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	if err := c.storeTx(ctx, txHash); err != nil {
		return common.Hash{}, err
	}

	return txHash, nil
}
