package postage

import "math/big"

type Reservestate struct {
	Radius    uint8    `json:"radius"`
	Available int64    `json:"available"`
	Outer     *big.Int `json:"outer"` // lower value limit for outer layer = the further half of chunks
	Inner     *big.Int `json:"inner"`
}
