package pledge

import (
	"context"
	"errors"
	hopabi "github.com/redesblock/hop/contracts/abi"
	"github.com/redesblock/hop/core/settlement/swap/erc20"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/redesblock/hop/core/sctx"
	"github.com/redesblock/hop/core/transaction"
)

var (
	pledgeABI            = transaction.ParseABIUnchecked(hopabi.PledgepABI)
	errDecodeABI         = errors.New("could not decode abi data")
	ErrInsufficientFunds = errors.New("insufficient token balance")
)

type Service interface {
	Stake(ctx context.Context, value *big.Int) (common.Hash, error)
	UnStake(ctx context.Context, value *big.Int) (common.Hash, error)
	GetShare(ctx context.Context, address common.Address) (*big.Int, error)
	GetSlash(ctx context.Context, address common.Address) (*big.Int, error)
	GetTotalShare(ctx context.Context) (*big.Int, error)
	GetTotalSlash(ctx context.Context) (*big.Int, error)
}

type pledgeService struct {
	overlayEthAddress  common.Address
	backend            transaction.Backend
	transactionService transaction.Service
	address            common.Address
}

func New(overlayEthAddress common.Address, backend transaction.Backend, transactionService transaction.Service, address common.Address) Service {
	return &pledgeService{
		overlayEthAddress:  overlayEthAddress,
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

func (c *pledgeService) GetTotalShare(ctx context.Context) (*big.Int, error) {
	callData, err := pledgeABI.Pack("totalShare")
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

	results, err := pledgeABI.Unpack("totalShare", output)
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

func (c *pledgeService) GetTotalSlash(ctx context.Context) (*big.Int, error) {
	callData, err := pledgeABI.Pack("totalSlash")
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

	results, err := pledgeABI.Unpack("totalSlash", output)
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
	erc20Address, err := c.ERC20Address(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	erc20Service := erc20.New(c.backend, c.transactionService, erc20Address)
	balance, err := erc20Service.BalanceOf(ctx, c.overlayEthAddress)
	if err != nil {
		return common.Hash{}, err
	}

	// check we can afford this so we don't waste gas and don't risk bouncing cheques
	if balance.Cmp(value) < 0 {
		return common.Hash{}, ErrInsufficientFunds
	}

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

	receipt, err := c.transactionService.WaitForReceipt(ctx, txHash)
	if err != nil {
		return common.Hash{}, err
	}

	if receipt.Status == 0 {
		return common.Hash{}, transaction.ErrTransactionReverted
	}

	return txHash, nil
}

func (c *pledgeService) UnStake(ctx context.Context, value *big.Int) (common.Hash, error) {
	stakedBalance, err := c.GetShare(ctx, c.overlayEthAddress)
	if err != nil {
		return common.Hash{}, err
	}

	// check we can afford this so we don't waste gas and don't risk bouncing cheques
	if stakedBalance.Cmp(value) < 0 {
		return common.Hash{}, ErrInsufficientFunds
	}

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

	receipt, err := c.transactionService.WaitForReceipt(ctx, txHash)
	if err != nil {
		return common.Hash{}, err
	}

	if receipt.Status == 0 {
		return common.Hash{}, transaction.ErrTransactionReverted
	}

	return txHash, nil
}

func (c *pledgeService) ERC20Address(ctx context.Context) (common.Address, error) {
	callData, err := pledgeABI.Pack("stakeToken")
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

	results, err := pledgeABI.Unpack("stakeToken", output)
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
