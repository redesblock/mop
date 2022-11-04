package bigint_test

import (
	"encoding/json"
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/redesblock/mop/core/util/bigint"
)

func TestMarshaling(t *testing.T) {
	mar, err := json.Marshal(struct {
		Bg *bigint.BigInt
	}{
		Bg: bigint.Wrap(new(big.Int).Mul(big.NewInt(math.MaxInt64), big.NewInt(math.MaxInt64))),
	})
	if err != nil {
		t.Errorf("Marshaling failed: %v", err)
	}
	if !reflect.DeepEqual(mar, []byte("{\"Bg\":\"85070591730234615847396907784232501249\"}")) {
		t.Error("Wrongly marshaled data")
	}
}
