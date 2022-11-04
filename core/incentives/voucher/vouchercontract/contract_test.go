package vouchercontract_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/redesblock/mop/core/chain/transaction"
	transactionMock "github.com/redesblock/mop/core/chain/transaction/mock"
	"github.com/redesblock/mop/core/incentives/voucher"
	voucherstoreMock "github.com/redesblock/mop/core/incentives/voucher/batchstore/mock"
	voucherMock "github.com/redesblock/mop/core/incentives/voucher/mock"
	vouchertesting "github.com/redesblock/mop/core/incentives/voucher/testing"
	"github.com/redesblock/mop/core/incentives/voucher/vouchercontract"
)

func TestCreateBatch(t *testing.T) {
	defer func(b uint8) {
		vouchercontract.BucketDepth = b
	}(vouchercontract.BucketDepth)
	vouchercontract.BucketDepth = 9
	owner := common.HexToAddress("abcd")
	label := "label"
	voucherStampAddress := common.HexToAddress("ffff")
	mopTokenAddress := common.HexToAddress("eeee")
	ctx := context.Background()
	initialBalance := big.NewInt(100)

	t.Run("ok", func(t *testing.T) {

		depth := uint8(10)
		totalAmount := big.NewInt(102400)
		txHashApprove := common.HexToHash("abb0")
		txHashCreate := common.HexToHash("c3a7")
		batchID := common.HexToHash("dddd")
		voucherMock := voucherMock.New()

		expectedCallData, err := vouchercontract.VoucherStampABI.Pack("createBatch", owner, initialBalance, depth, vouchercontract.BucketDepth, common.Hash{}, false)
		if err != nil {
			t.Fatal(err)
		}

		contract := vouchercontract.New(
			owner,
			voucherStampAddress,
			mopTokenAddress,
			transactionMock.New(
				transactionMock.WithSendFunc(func(ctx context.Context, request *transaction.TxRequest) (txHash common.Hash, err error) {
					if *request.To == mopTokenAddress {
						return txHashApprove, nil
					} else if *request.To == voucherStampAddress {
						if !bytes.Equal(expectedCallData[:100], request.Data[:100]) {
							return common.Hash{}, fmt.Errorf("got wrong call data. wanted %x, got %x", expectedCallData, request.Data)
						}
						return txHashCreate, nil
					}
					return common.Hash{}, errors.New("sent to wrong contract")
				}),
				transactionMock.WithWaitForReceiptFunc(func(ctx context.Context, txHash common.Hash) (receipt *types.Receipt, err error) {
					if txHash == txHashApprove {
						return &types.Receipt{
							Status: 1,
						}, nil
					} else if txHash == txHashCreate {
						return &types.Receipt{
							Logs: []*types.Log{
								newCreateEvent(voucherStampAddress, batchID),
							},
							Status: 1,
						}, nil
					}
					return nil, errors.New("unknown tx hash")
				}),
				transactionMock.WithCallFunc(func(ctx context.Context, request *transaction.TxRequest) (result []byte, err error) {
					if *request.To == mopTokenAddress {
						return totalAmount.FillBytes(make([]byte, 32)), nil
					}
					return nil, errors.New("unexpected call")
				}),
			),
			voucherMock,
			voucherstoreMock.New(),
			true,
		)

		returnedID, err := contract.CreateBatch(ctx, initialBalance, depth, false, label)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(returnedID, batchID[:]) {
			t.Fatalf("got wrong batchId. wanted %v, got %v", batchID, returnedID)
		}

		si, err := voucherMock.GetStampIssuer(returnedID)
		if err != nil {
			t.Fatal(err)
		}

		if si == nil {
			t.Fatal("stamp issuer not set")
		}
	})

	t.Run("invalid depth", func(t *testing.T) {
		depth := uint8(9)

		contract := vouchercontract.New(
			owner,
			voucherStampAddress,
			mopTokenAddress,
			transactionMock.New(),
			voucherMock.New(),
			voucherstoreMock.New(),
			true,
		)

		_, err := contract.CreateBatch(ctx, initialBalance, depth, false, label)
		if !errors.Is(err, vouchercontract.ErrInvalidDepth) {
			t.Fatalf("expected error %v. got %v", vouchercontract.ErrInvalidDepth, err)
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		depth := uint8(10)
		totalAmount := big.NewInt(102399)

		contract := vouchercontract.New(
			owner,
			voucherStampAddress,
			mopTokenAddress,
			transactionMock.New(
				transactionMock.WithCallFunc(func(ctx context.Context, request *transaction.TxRequest) (result []byte, err error) {
					if *request.To == mopTokenAddress {
						return big.NewInt(0).Sub(totalAmount, big.NewInt(1)).FillBytes(make([]byte, 32)), nil
					}
					return nil, errors.New("unexpected call")
				}),
			),
			voucherMock.New(),
			voucherstoreMock.New(),
			true,
		)

		_, err := contract.CreateBatch(ctx, initialBalance, depth, false, label)
		if !errors.Is(err, vouchercontract.ErrInsufficientFunds) {
			t.Fatalf("expected error %v. got %v", vouchercontract.ErrInsufficientFunds, err)
		}
	})
}

func newCreateEvent(voucherContractAddress common.Address, batchId common.Hash) *types.Log {
	b, err := vouchercontract.VoucherStampABI.Events["BatchCreated"].Inputs.NonIndexed().Pack(
		big.NewInt(0),
		big.NewInt(0),
		common.Address{},
		uint8(1),
		uint8(2),
		false,
	)
	if err != nil {
		panic(err)
	}
	return &types.Log{
		Address: voucherContractAddress,
		Data:    b,
		Topics:  []common.Hash{vouchercontract.BatchCreatedTopic, batchId},
	}
}

func TestLookupERC20Address(t *testing.T) {
	voucherStampAddress := common.HexToAddress("ffff")
	erc20Address := common.HexToAddress("ffff")

	addr, err := vouchercontract.LookupERC20Address(
		context.Background(),
		transactionMock.New(
			transactionMock.WithCallFunc(func(ctx context.Context, request *transaction.TxRequest) (result []byte, err error) {
				if *request.To != voucherStampAddress {
					return nil, fmt.Errorf("called wrong contract. wanted %v, got %v", voucherStampAddress, request.To)
				}
				return erc20Address.Hash().Bytes(), nil
			}),
		),
		voucherStampAddress,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}

	if addr != voucherStampAddress {
		t.Fatalf("got wrong erc20 address. wanted %v, got %v", erc20Address, addr)
	}
}

func TestTopUpBatch(t *testing.T) {
	defer func(b uint8) {
		vouchercontract.BucketDepth = b
	}(vouchercontract.BucketDepth)
	vouchercontract.BucketDepth = 9
	owner := common.HexToAddress("abcd")
	voucherStampAddress := common.HexToAddress("ffff")
	mopTokenAddress := common.HexToAddress("eeee")
	ctx := context.Background()
	topupBalance := big.NewInt(100)

	t.Run("ok", func(t *testing.T) {

		totalAmount := big.NewInt(102400)
		txHashApprove := common.HexToHash("abb0")
		txHashTopup := common.HexToHash("c3a7")
		batch := vouchertesting.MustNewBatch(vouchertesting.WithOwner(owner.Bytes()))
		batch.Depth = uint8(10)
		batch.BucketDepth = uint8(9)
		voucherMock := voucherMock.New(voucherMock.WithIssuer(voucher.NewStampIssuer(
			"label",
			"keyID",
			batch.ID,
			batch.Value,
			batch.Depth,
			batch.BucketDepth,
			batch.Start,
			batch.Immutable,
		)))
		batchStoreMock := voucherstoreMock.New(voucherstoreMock.WithBatch(batch))

		expectedCallData, err := vouchercontract.VoucherStampABI.Pack("topUp", common.BytesToHash(batch.ID), topupBalance)
		if err != nil {
			t.Fatal(err)
		}

		contract := vouchercontract.New(
			owner,
			voucherStampAddress,
			mopTokenAddress,
			transactionMock.New(
				transactionMock.WithSendFunc(func(ctx context.Context, request *transaction.TxRequest) (txHash common.Hash, err error) {
					if *request.To == mopTokenAddress {
						return txHashApprove, nil
					} else if *request.To == voucherStampAddress {
						if !bytes.Equal(expectedCallData[:64], request.Data[:64]) {
							return common.Hash{}, fmt.Errorf("got wrong call data. wanted %x, got %x", expectedCallData, request.Data)
						}
						return txHashTopup, nil
					}
					return common.Hash{}, errors.New("sent to wrong contract")
				}),
				transactionMock.WithWaitForReceiptFunc(func(ctx context.Context, txHash common.Hash) (receipt *types.Receipt, err error) {
					if txHash == txHashApprove {
						return &types.Receipt{
							Status: 1,
						}, nil
					} else if txHash == txHashTopup {
						return &types.Receipt{
							Logs: []*types.Log{
								newTopUpEvent(voucherStampAddress, batch),
							},
							Status: 1,
						}, nil
					}
					return nil, errors.New("unknown tx hash")
				}),
				transactionMock.WithCallFunc(func(ctx context.Context, request *transaction.TxRequest) (result []byte, err error) {
					if *request.To == mopTokenAddress {
						return totalAmount.FillBytes(make([]byte, 32)), nil
					}
					return nil, errors.New("unexpected call")
				}),
			),
			voucherMock,
			batchStoreMock,
			true,
		)

		err = contract.TopUpBatch(ctx, batch.ID, topupBalance)
		if err != nil {
			t.Fatal(err)
		}

		si, err := voucherMock.GetStampIssuer(batch.ID)
		if err != nil {
			t.Fatal(err)
		}

		if si == nil {
			t.Fatal("stamp issuer not set")
		}
	})

	t.Run("batch doesnt exist", func(t *testing.T) {
		errNotFound := errors.New("not found")
		contract := vouchercontract.New(
			owner,
			voucherStampAddress,
			mopTokenAddress,
			transactionMock.New(),
			voucherMock.New(),
			voucherstoreMock.New(voucherstoreMock.WithGetErr(errNotFound, 0)),
			true,
		)

		err := contract.TopUpBatch(ctx, vouchertesting.MustNewID(), topupBalance)
		if !errors.Is(err, errNotFound) {
			t.Fatal("expected error on topup of non existent batch")
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		totalAmount := big.NewInt(102399)
		batch := vouchertesting.MustNewBatch(vouchertesting.WithOwner(owner.Bytes()))
		batchStoreMock := voucherstoreMock.New(voucherstoreMock.WithBatch(batch))

		contract := vouchercontract.New(
			owner,
			voucherStampAddress,
			mopTokenAddress,
			transactionMock.New(
				transactionMock.WithCallFunc(func(ctx context.Context, request *transaction.TxRequest) (result []byte, err error) {
					if *request.To == mopTokenAddress {
						return big.NewInt(0).Sub(totalAmount, big.NewInt(1)).FillBytes(make([]byte, 32)), nil
					}
					return nil, errors.New("unexpected call")
				}),
			),
			voucherMock.New(),
			batchStoreMock,
			true,
		)

		err := contract.TopUpBatch(ctx, batch.ID, topupBalance)
		if !errors.Is(err, vouchercontract.ErrInsufficientFunds) {
			t.Fatalf("expected error %v. got %v", vouchercontract.ErrInsufficientFunds, err)
		}
	})
}

func newTopUpEvent(voucherContractAddress common.Address, batch *voucher.Batch) *types.Log {
	b, err := vouchercontract.VoucherStampABI.Events["BatchTopUp"].Inputs.NonIndexed().Pack(
		big.NewInt(0),
		big.NewInt(0),
	)
	if err != nil {
		panic(err)
	}
	return &types.Log{
		Address:     voucherContractAddress,
		Data:        b,
		Topics:      []common.Hash{vouchercontract.BatchTopUpTopic, common.BytesToHash(batch.ID)},
		BlockNumber: batch.Start + 1,
	}
}

func TestDiluteBatch(t *testing.T) {
	defer func(b uint8) {
		vouchercontract.BucketDepth = b
	}(vouchercontract.BucketDepth)
	vouchercontract.BucketDepth = 9
	owner := common.HexToAddress("abcd")
	voucherStampAddress := common.HexToAddress("ffff")
	mopTokenAddress := common.HexToAddress("eeee")
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {

		txHashDilute := common.HexToHash("c3a7")
		batch := vouchertesting.MustNewBatch(vouchertesting.WithOwner(owner.Bytes()))
		batch.Depth = uint8(10)
		batch.BucketDepth = uint8(9)
		batch.Value = big.NewInt(100)
		newDepth := batch.Depth + 1
		voucherMock := voucherMock.New(voucherMock.WithIssuer(voucher.NewStampIssuer(
			"label",
			"keyID",
			batch.ID,
			batch.Value,
			batch.Depth,
			batch.BucketDepth,
			batch.Start,
			batch.Immutable,
		)))
		batchStoreMock := voucherstoreMock.New(voucherstoreMock.WithBatch(batch))

		expectedCallData, err := vouchercontract.VoucherStampABI.Pack("increaseDepth", common.BytesToHash(batch.ID), newDepth)
		if err != nil {
			t.Fatal(err)
		}

		contract := vouchercontract.New(
			owner,
			voucherStampAddress,
			mopTokenAddress,
			transactionMock.New(
				transactionMock.WithSendFunc(func(ctx context.Context, request *transaction.TxRequest) (txHash common.Hash, err error) {
					if *request.To == voucherStampAddress {
						if !bytes.Equal(expectedCallData[:64], request.Data[:64]) {
							return common.Hash{}, fmt.Errorf("got wrong call data. wanted %x, got %x", expectedCallData, request.Data)
						}
						return txHashDilute, nil
					}
					return common.Hash{}, errors.New("sent to wrong contract")
				}),
				transactionMock.WithWaitForReceiptFunc(func(ctx context.Context, txHash common.Hash) (receipt *types.Receipt, err error) {
					if txHash == txHashDilute {
						return &types.Receipt{
							Logs: []*types.Log{
								newDiluteEvent(voucherStampAddress, batch),
							},
							Status: 1,
						}, nil
					}
					return nil, errors.New("unknown tx hash")
				}),
			),
			voucherMock,
			batchStoreMock,
			true,
		)

		err = contract.DiluteBatch(ctx, batch.ID, newDepth)
		if err != nil {
			t.Fatal(err)
		}

		si, err := voucherMock.GetStampIssuer(batch.ID)
		if err != nil {
			t.Fatal(err)
		}

		if si == nil {
			t.Fatal("stamp issuer not set")
		}
	})

	t.Run("batch doesnt exist", func(t *testing.T) {
		errNotFound := errors.New("not found")
		contract := vouchercontract.New(
			owner,
			voucherStampAddress,
			mopTokenAddress,
			transactionMock.New(),
			voucherMock.New(),
			voucherstoreMock.New(voucherstoreMock.WithGetErr(errNotFound, 0)),
			true,
		)

		err := contract.DiluteBatch(ctx, vouchertesting.MustNewID(), uint8(17))
		if !errors.Is(err, errNotFound) {
			t.Fatal("expected error on topup of non existent batch")
		}
	})

	t.Run("invalid depth", func(t *testing.T) {
		batch := vouchertesting.MustNewBatch(vouchertesting.WithOwner(owner.Bytes()))
		batch.Depth = uint8(16)
		batchStoreMock := voucherstoreMock.New(voucherstoreMock.WithBatch(batch))

		contract := vouchercontract.New(
			owner,
			voucherStampAddress,
			mopTokenAddress,
			transactionMock.New(),
			voucherMock.New(),
			batchStoreMock,
			true,
		)

		err := contract.DiluteBatch(ctx, batch.ID, batch.Depth-1)
		if !errors.Is(err, vouchercontract.ErrInvalidDepth) {
			t.Fatalf("expected error %v. got %v", vouchercontract.ErrInvalidDepth, err)
		}
	})
}

func newDiluteEvent(voucherContractAddress common.Address, batch *voucher.Batch) *types.Log {
	b, err := vouchercontract.VoucherStampABI.Events["BatchDepthIncrease"].Inputs.NonIndexed().Pack(
		uint8(0),
		big.NewInt(0),
	)
	if err != nil {
		panic(err)
	}
	return &types.Log{
		Address:     voucherContractAddress,
		Data:        b,
		Topics:      []common.Hash{vouchercontract.BatchDiluteTopic, common.BytesToHash(batch.ID)},
		BlockNumber: batch.Start + 1,
	}
}
