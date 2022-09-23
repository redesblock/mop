package erc20

import (
	"context"
	"errors"
	mopabi "github.com/redesblock/mop/contracts/abi"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/redesblock/mop/core/sctx"
	"github.com/redesblock/mop/core/transaction"
)

var (
	erc20ABI     = transaction.ParseABIUnchecked(mopabi.ERC20ABI)
	errDecodeABI = errors.New("could not decode abi data")
)

type Service interface {
	TotalSupply(ctx context.Context) (*big.Int, error)
	BalanceOf(ctx context.Context, address common.Address) (*big.Int, error)
	Transfer(ctx context.Context, address common.Address, value *big.Int) (common.Hash, error)
	Approval(ctx context.Context, spender common.Address, value *big.Int) (common.Hash, error)
	Allowance(ctx context.Context, owner common.Address, spender common.Address) (*big.Int, error)
}

type erc20Service struct {
	backend            transaction.Backend
	transactionService transaction.Service
	address            common.Address
}

func New(backend transaction.Backend, transactionService transaction.Service, address common.Address) Service {
	return &erc20Service{
		backend:            backend,
		transactionService: transactionService,
		address:            address,
	}
}

func (c *erc20Service) TotalSupply(ctx context.Context) (*big.Int, error) {
	callData, err := erc20ABI.Pack("totalSupply")
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

	results, err := erc20ABI.Unpack("totalSupply", output)
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

func (c *erc20Service) BalanceOf(ctx context.Context, address common.Address) (*big.Int, error) {
	callData, err := erc20ABI.Pack("balanceOf", address)
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

	results, err := erc20ABI.Unpack("balanceOf", output)
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

func (c *erc20Service) Transfer(ctx context.Context, address common.Address, value *big.Int) (common.Hash, error) {
	callData, err := erc20ABI.Pack("transfer", address, value)
	if err != nil {
		return common.Hash{}, err
	}

	request := &transaction.TxRequest{
		To:          &c.address,
		Data:        callData,
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    90000,
		Value:       big.NewInt(0),
		Description: "token transfer",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	return txHash, nil
}

func (c *erc20Service) Allowance(ctx context.Context, owner common.Address, spender common.Address) (*big.Int, error) {
	callData, err := erc20ABI.Pack("allowance", owner, spender)
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

	results, err := erc20ABI.Unpack("allowance", output)
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

func (c *erc20Service) Approval(ctx context.Context, spender common.Address, value *big.Int) (common.Hash, error) {
	callData, err := erc20ABI.Pack("approve", spender, value)
	if err != nil {
		return common.Hash{}, err
	}

	request := &transaction.TxRequest{
		To:          &c.address,
		Data:        callData,
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    90000,
		Value:       big.NewInt(0),
		Description: "token spender",
	}

	txHash, err := c.transactionService.Send(ctx, request)
	if err != nil {
		return common.Hash{}, err
	}

	return txHash, nil
}
