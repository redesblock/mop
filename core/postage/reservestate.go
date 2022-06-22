package postage

import "math/big"

type ReserveState struct {
	Radius    uint8
	Available int64
	Outer     *big.Int // lower value limit for outer layer = the further half of chunks
	Inner     *big.Int
}
