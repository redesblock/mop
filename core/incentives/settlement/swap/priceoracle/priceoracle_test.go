package priceoracle_test

import (
	"context"
	mabi "github.com/redesblock/mop/core/contract/abi"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/redesblock/mop/core/chain/transaction"
	transactionmock "github.com/redesblock/mop/core/chain/transaction/mock"
	"github.com/redesblock/mop/core/incentives/settlement/swap/priceoracle"
	"github.com/redesblock/mop/core/log"
)

var (
	priceOracleABI = transaction.ParseABIUnchecked(mabi.PriceOracleABIv0_1_0)
)

func TestExchangeGetPrice(t *testing.T) {
	priceOracleAddress := common.HexToAddress("0xabcd")

	expectedPrice := big.NewInt(100)
	expectedDeduce := big.NewInt(200)

	result := make([]byte, 64)
	expectedPrice.FillBytes(result[0:32])
	expectedDeduce.FillBytes(result[32:64])

	ex := priceoracle.New(
		log.Noop,
		priceOracleAddress,
		transactionmock.New(
			transactionmock.WithABICall(
				&priceOracleABI,
				priceOracleAddress,
				result,
				"getPrice",
			),
		),
		1,
	)

	price, deduce, err := ex.GetPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if expectedPrice.Cmp(price) != 0 {
		t.Fatalf("got wrong price. wanted %d, got %d", expectedPrice, price)
	}

	if expectedDeduce.Cmp(deduce) != 0 {
		t.Fatalf("got wrong deduce. wanted %d, got %d", expectedDeduce, deduce)
	}
}
