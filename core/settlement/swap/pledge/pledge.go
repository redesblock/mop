package pledge

import (
	"context"
	"errors"
	hopabi "github.com/redesblock/hop/contracts/abi"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/redesblock/hop/core/sctx"
	"github.com/redesblock/hop/core/transaction"
)

var (
	pledgeABI    = transaction.ParseABIUnchecked(hopabi.PledgepABI)
	errDecodeABI = errors.New("could not decode abi data")
)

type Service interface {
	Stake(ctx context.Context, value *big.Int) (common.Hash, error)
	UnStake(ctx context.Context, value *big.Int) (common.Hash, error)
	GetShare(ctx context.Context, address common.Address) (*big.Int, error)
	GetSlash(ctx context.Context, address common.Address) (*big.Int, error)
}

type pledgeService struct {
	backend            transaction.Backend
	transactionService transaction.Service
	address            common.Address
}

func New(backend transaction.Backend, transactionService transaction.Service, address common.Address) Service {
	return &pledgeService{
		backend:            backend,
		transactionService: transactionService,
		address:            address,
	}
}

func (c *pledgeService) GetShare(ctx context.Context, address common.Address) (*big.Int, error) {
	callData, err := pledgeABI.Pack("getShare", address)
	if err != nil {
		return nil, err
	}

	output, err := c.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &c.address,
		Data: callData,
	})
	if err != nil {
		return nil, err
	}

	results, err := pledgeABI.Unpack("getShare", output)
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

func (c *pledgeService) GetSlash(ctx context.Context, address common.Address) (*big.Int, error) {
	callData, err := pledgeABI.Pack("getSlash", address)
	if err != nil {
		return nil, err
	}

	output, err := c.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &c.address,
		Data: callData,
	})
	if err != nil {
		return nil, err
	}

	results, err := pledgeABI.Unpack("getSlash", output)
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

func (c *pledgeService) Stake(ctx context.Context, value *big.Int) (common.Hash, error) {
	callData, err := pledgeABI.Pack("stake", value)
	if err != nil {
		return common.Hash{}, err
	}

	request := &transaction.TxRequest{
		To:          &c.address,
		Data:        callData,
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    90000,
		Value:       big.NewInt(0),
		Description: "token stake",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	return txHash, nil
}

func (c *pledgeService) UnStake(ctx context.Context, value *big.Int) (common.Hash, error) {
	callData, err := pledgeABI.Pack("unStake", value)
	if err != nil {
		return common.Hash{}, err
	}

	request := &transaction.TxRequest{
		To:          &c.address,
		Data:        callData,
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    90000,
		Value:       big.NewInt(0),
		Description: "token unstake",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	return txHash, nil
}
