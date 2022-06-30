package pledge

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	hopabi "github.com/redesblock/hop/contracts/abi"
	"github.com/redesblock/hop/core/settlement/swap/erc20"
	"github.com/redesblock/hop/core/storage"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/redesblock/hop/core/sctx"
	"github.com/redesblock/hop/core/transaction"
)

var (
	erc20ABI             = transaction.ParseABIUnchecked(hopabi.ERC20ABI)
	pledgeABI            = transaction.ParseABIUnchecked(hopabi.PledgepABI)
	errDecodeABI         = errors.New("could not decode abi data")
	ErrInsufficientFunds = errors.New("insufficient token balance")
	keyPrefix            = "pledge-txs-"
)

type Service interface {
	Stake(ctx context.Context, address common.Address, value *big.Int) (common.Hash, error)
	UnStake(ctx context.Context, address common.Address, value *big.Int) (common.Hash, error)
	GetShare(ctx context.Context, address common.Address) (*big.Int, error)
	GetSlash(ctx context.Context, address common.Address) (*big.Int, error)
	GetTotalShare(ctx context.Context) (*big.Int, error)
	GetTotalSlash(ctx context.Context) (*big.Int, error)
	AvailableBalance(ctx context.Context, address common.Address) (*big.Int, error)
	Txs() ([]string, error)
}

type pledgeService struct {
	stateStore         storage.StateStorer
	backend            transaction.Backend
	transactionService transaction.Service
	address            common.Address
}

func New(stateStore storage.StateStorer, backend transaction.Backend, transactionService transaction.Service, address common.Address) Service {
	return &pledgeService{
		stateStore:         stateStore,
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

func (c *pledgeService) Stake(ctx context.Context, address common.Address, value *big.Int) (common.Hash, error) {
	balance, err := c.AvailableBalance(ctx, address)
	if err != nil {
		return common.Hash{}, err
	}

	// check we can afford this so we don't waste gas and don't risk bouncing cheques
	if balance.Cmp(value) < 0 {
		return common.Hash{}, ErrInsufficientFunds
	}

	if _, err := c.sendApproveTransaction(ctx, value); err != nil {
		return common.Hash{}, err
	}

	callData, err := pledgeABI.Pack("stake", value)
	if err != nil {
		return common.Hash{}, err
	}

	request := &transaction.TxRequest{
		To:          &c.address,
		Data:        callData,
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    100000,
		Value:       big.NewInt(0),
		Description: "token stake",
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

func (c *pledgeService) AvailableBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	erc20Address, err := c.LookupERC20Address(ctx)
	if err != nil {
		return nil, err
	}

	erc20Service := erc20.New(c.backend, c.transactionService, erc20Address)
	balance, err := erc20Service.BalanceOf(ctx, address)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (c *pledgeService) UnStake(ctx context.Context, address common.Address, value *big.Int) (common.Hash, error) {
	stakedBalance, err := c.GetShare(ctx, address)
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
		GasLimit:    100000,
		Value:       big.NewInt(0),
		Description: "token unstake",
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

func (c *pledgeService) LookupERC20Address(ctx context.Context) (common.Address, error) {
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

func (c *pledgeService) Txs() ([]string, error) {
	var txs []string
	if err := c.stateStore.Iterate(keyPrefix, func(k, v []byte) (bool, error) {
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

func (c *pledgeService) storeTx(ctx context.Context, txHash common.Hash) error {
	receipt, err := c.transactionService.WaitForReceipt(ctx, txHash)
	if err != nil {
		return err
	}

	if c.stateStore != nil {
		c.stateStore.Put(keyPrefix+txHash.String(), receipt)
	}

	if receipt.Status == 0 {
		return transaction.ErrTransactionReverted
	}
	return nil
}

func (c *pledgeService) sendApproveTransaction(ctx context.Context, amount *big.Int) (*types.Receipt, error) {
	erc20Address, err := c.LookupERC20Address(ctx)
	if err != nil {
		return nil, err
	}
	callData, err := erc20ABI.Pack("approve", c.address, amount)
	if err != nil {
		return nil, err
	}

	txHash, err := c.transactionService.Send(ctx, &transaction.TxRequest{
		To:          &erc20Address,
		Data:        callData,
		GasPrice:    sctx.GetGasPrice(ctx),
		GasLimit:    65000,
		Value:       big.NewInt(0),
		Description: "Approve tokens for pledge operations",
	})
	if err != nil {
		return nil, err
	}

	receipt, err := c.transactionService.WaitForReceipt(ctx, txHash)
	if err != nil {
		return nil, err
	}

	if receipt.Status == 0 {
		return nil, transaction.ErrTransactionReverted
	}

	return receipt, nil
}
