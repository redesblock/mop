package vouchercontract

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	mabi "github.com/redesblock/mop/core/contract/abi"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/redesblock/mop/core/chain/transaction"
	"github.com/redesblock/mop/core/incentives/voucher"
	"github.com/redesblock/mop/core/mctx"
)

var (
	BucketDepth = uint8(16)

	voucherStampABI   = parseABI(mabi.VoucherStampABIv0_1_0)
	erc20ABI          = parseABI(mabi.ERC20ABIv0_1_0)
	batchCreatedTopic = voucherStampABI.Events["BatchCreated"].ID
	batchTopUpTopic   = voucherStampABI.Events["BatchTopUp"].ID
	batchDiluteTopic  = voucherStampABI.Events["BatchDepthIncrease"].ID

	ErrBatchCreate       = errors.New("batch creation failed")
	ErrInsufficientFunds = errors.New("insufficient token balance")
	ErrInvalidDepth      = errors.New("invalid depth")
	ErrBatchTopUp        = errors.New("batch topUp failed")
	ErrBatchDilute       = errors.New("batch dilute failed")
	ErrChainDisabled     = errors.New("chain disabled")

	approveDescription     = "Approve tokens for voucher operations"
	createBatchDescription = "Voucher batch creation"
	topUpBatchDescription  = "Voucher batch top up"
	diluteBatchDescription = "Voucher batch dilute"
)

type Interface interface {
	CreateBatch(ctx context.Context, initialBalance *big.Int, depth uint8, immutable bool, label string) ([]byte, error)
	TopUpBatch(ctx context.Context, batchID []byte, topupBalance *big.Int) error
	DiluteBatch(ctx context.Context, batchID []byte, newDepth uint8) error
}

type voucherContract struct {
	owner                  common.Address
	voucherContractAddress common.Address
	mopTokenAddress        common.Address
	transactionService     transaction.Service
	voucherService         voucher.Service
	voucherStorer          voucher.Storer
}

func New(
	owner,
	voucherContractAddress,
	mopTokenAddress common.Address,
	transactionService transaction.Service,
	voucherService voucher.Service,
	voucherStorer voucher.Storer,
	chainEnabled bool,
) Interface {
	if !chainEnabled {
		return new(noOpVoucherContract)
	}

	return &voucherContract{
		owner:                  owner,
		voucherContractAddress: voucherContractAddress,
		mopTokenAddress:        mopTokenAddress,
		transactionService:     transactionService,
		voucherService:         voucherService,
		voucherStorer:          voucherStorer,
	}
}

func (c *voucherContract) sendApproveTransaction(ctx context.Context, amount *big.Int) (*types.Receipt, error) {
	callData, err := erc20ABI.Pack("approve", c.voucherContractAddress, amount)
	if err != nil {
		return nil, err
	}

	txHash, err := c.transactionService.Send(ctx, &transaction.TxRequest{
		To:          &c.mopTokenAddress,
		Data:        callData,
		GasPrice:    mctx.GetGasPrice(ctx),
		GasLimit:    65000,
		Value:       big.NewInt(0),
		Description: approveDescription,
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

func (c *voucherContract) sendTransaction(ctx context.Context, callData []byte, desc string) (*types.Receipt, error) {
	request := &transaction.TxRequest{
		To:          &c.voucherContractAddress,
		Data:        callData,
		GasPrice:    mctx.GetGasPrice(ctx),
		GasLimit:    1600000,
		Value:       big.NewInt(0),
		Description: desc,
	}

	txHash, err := c.transactionService.Send(ctx, request)
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

func (c *voucherContract) sendCreateBatchTransaction(ctx context.Context, owner common.Address, initialBalance *big.Int, depth uint8, nonce common.Hash, immutable bool) (*types.Receipt, error) {

	callData, err := voucherStampABI.Pack("createBatch", owner, initialBalance, depth, BucketDepth, nonce, immutable)
	if err != nil {
		return nil, err
	}

	receipt, err := c.sendTransaction(ctx, callData, createBatchDescription)
	if err != nil {
		return nil, fmt.Errorf("create batch: depth %d bucketDepth %d immutable %t: %w", depth, BucketDepth, immutable, err)
	}

	return receipt, nil
}

func (c *voucherContract) sendTopUpBatchTransaction(ctx context.Context, batchID []byte, topUpAmount *big.Int) (*types.Receipt, error) {

	callData, err := voucherStampABI.Pack("topUp", common.BytesToHash(batchID), topUpAmount)
	if err != nil {
		return nil, err
	}

	receipt, err := c.sendTransaction(ctx, callData, topUpBatchDescription)
	if err != nil {
		return nil, fmt.Errorf("topup batch: amount %d: %w", topUpAmount.Int64(), err)
	}

	return receipt, nil
}

func (c *voucherContract) sendDiluteTransaction(ctx context.Context, batchID []byte, newDepth uint8) (*types.Receipt, error) {

	callData, err := voucherStampABI.Pack("increaseDepth", common.BytesToHash(batchID), newDepth)
	if err != nil {
		return nil, err
	}

	receipt, err := c.sendTransaction(ctx, callData, diluteBatchDescription)
	if err != nil {
		return nil, fmt.Errorf("dilute batch: new depth %d: %w", newDepth, err)
	}

	return receipt, nil
}

func (c *voucherContract) getBalance(ctx context.Context) (*big.Int, error) {
	callData, err := erc20ABI.Pack("balanceOf", c.owner)
	if err != nil {
		return nil, err
	}

	result, err := c.transactionService.Call(ctx, &transaction.TxRequest{
		To:   &c.mopTokenAddress,
		Data: callData,
	})
	if err != nil {
		return nil, err
	}

	results, err := erc20ABI.Unpack("balanceOf", result)
	if err != nil {
		return nil, err
	}
	return abi.ConvertType(results[0], new(big.Int)).(*big.Int), nil
}

func (c *voucherContract) CreateBatch(ctx context.Context, initialBalance *big.Int, depth uint8, immutable bool, label string) ([]byte, error) {

	if depth <= BucketDepth {
		return nil, ErrInvalidDepth
	}

	totalAmount := big.NewInt(0).Mul(initialBalance, big.NewInt(int64(1<<depth)))
	balance, err := c.getBalance(ctx)
	if err != nil {
		return nil, err
	}

	if balance.Cmp(totalAmount) < 0 {
		return nil, ErrInsufficientFunds
	}

	_, err = c.sendApproveTransaction(ctx, totalAmount)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 32)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	receipt, err := c.sendCreateBatchTransaction(ctx, c.owner, initialBalance, depth, common.BytesToHash(nonce), immutable)
	if err != nil {
		return nil, err
	}

	for _, ev := range receipt.Logs {
		if ev.Address == c.voucherContractAddress && len(ev.Topics) > 0 && ev.Topics[0] == batchCreatedTopic {
			var createdEvent batchCreatedEvent
			err = transaction.ParseEvent(&voucherStampABI, "BatchCreated", &createdEvent, *ev)
			if err != nil {
				return nil, err
			}

			batchID := createdEvent.BatchId[:]
			err = c.voucherService.Add(voucher.NewStampIssuer(
				label,
				c.owner.Hex(),
				batchID,
				initialBalance,
				createdEvent.Depth,
				createdEvent.BucketDepth,
				ev.BlockNumber,
				createdEvent.ImmutableFlag,
			))

			if err != nil {
				return nil, err
			}

			return createdEvent.BatchId[:], nil
		}
	}

	return nil, ErrBatchCreate
}

func (c *voucherContract) TopUpBatch(ctx context.Context, batchID []byte, topUpAmount *big.Int) error {

	batch, err := c.voucherStorer.Get(batchID)
	if err != nil {
		return err
	}

	totalAmount := big.NewInt(0).Mul(topUpAmount, big.NewInt(int64(1<<batch.Depth)))
	balance, err := c.getBalance(ctx)
	if err != nil {
		return err
	}

	if balance.Cmp(totalAmount) < 0 {
		return ErrInsufficientFunds
	}

	_, err = c.sendApproveTransaction(ctx, totalAmount)
	if err != nil {
		return err
	}

	receipt, err := c.sendTopUpBatchTransaction(ctx, batch.ID, topUpAmount)
	if err != nil {
		return err
	}

	for _, ev := range receipt.Logs {
		if ev.Address == c.voucherContractAddress && len(ev.Topics) > 0 && ev.Topics[0] == batchTopUpTopic {
			return nil
		}
	}

	return ErrBatchTopUp
}

func (c *voucherContract) DiluteBatch(ctx context.Context, batchID []byte, newDepth uint8) error {

	batch, err := c.voucherStorer.Get(batchID)
	if err != nil {
		return err
	}

	if batch.Depth > newDepth {
		return fmt.Errorf("new depth should be greater: %w", ErrInvalidDepth)
	}

	receipt, err := c.sendDiluteTransaction(ctx, batch.ID, newDepth)
	if err != nil {
		return err
	}

	for _, ev := range receipt.Logs {
		if ev.Address == c.voucherContractAddress && len(ev.Topics) > 0 && ev.Topics[0] == batchDiluteTopic {
			return nil
		}
	}

	return ErrBatchDilute
}

type batchCreatedEvent struct {
	BatchId           [32]byte
	TotalAmount       *big.Int
	NormalisedBalance *big.Int
	Owner             common.Address
	Depth             uint8
	BucketDepth       uint8
	ImmutableFlag     bool
}

func parseABI(json string) abi.ABI {
	cabi, err := abi.JSON(strings.NewReader(json))
	if err != nil {
		panic(fmt.Sprintf("error creating ABI for voucher contract: %v", err))
	}
	return cabi
}

func LookupERC20Address(ctx context.Context, transactionService transaction.Service, voucherContractAddress common.Address, chainEnabled bool) (common.Address, error) {
	if !chainEnabled {
		return common.Address{}, nil
	}

	callData, err := voucherStampABI.Pack("token")
	if err != nil {
		return common.Address{}, err
	}

	request := &transaction.TxRequest{
		To:       &voucherContractAddress,
		Data:     callData,
		GasPrice: nil,
		GasLimit: 0,
		Value:    big.NewInt(0),
	}

	data, err := transactionService.Call(ctx, request)
	if err != nil {
		return common.Address{}, err
	}

	return common.BytesToAddress(data), nil
}

type noOpVoucherContract struct{}

func (m *noOpVoucherContract) CreateBatch(context.Context, *big.Int, uint8, bool, string) ([]byte, error) {
	return nil, ErrChainDisabled
}
func (m *noOpVoucherContract) TopUpBatch(context.Context, []byte, *big.Int) error {
	return ErrChainDisabled
}
func (m *noOpVoucherContract) DiluteBatch(context.Context, []byte, uint8) error {
	return ErrChainDisabled
}
