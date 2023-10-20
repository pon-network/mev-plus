package common

import (
	"errors"
	"math/big"
	"os"
	"strconv"
)

type U256Str [32]byte

var (
	ErrSign   = errors.New("invalid sign")
	ErrLength = errors.New("invalid length")
)

const (
	SlotTimeSecMainnet = 12
)

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func GetEnvInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok {
		val, err := strconv.Atoi(value)
		if err == nil {
			return val
		}
	}
	return defaultValue
}

func GetEnvFloat64(key string, defaultValue float64) float64 {
	if value, ok := os.LookupEnv(key); ok {
		val, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return val
		}
	}
	return defaultValue
}

func (n *U256Str) FromBig(x *big.Int) error {
	if x.BitLen() > 256 {
		return ErrLength
	}
	if x.Sign() == -1 {
		return ErrSign
	}
	copy(n[:], reverse(x.FillBytes(n[:])))
	return nil
}

func reverse(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	for i := len(dst)/2 - 1; i >= 0; i-- {
		opp := len(dst) - 1 - i
		dst[i], dst[opp] = dst[opp], dst[i]
	}
	return dst
}

func (n *U256Str) BigInt() *big.Int {
	return new(big.Int).SetBytes(reverse(n[:]))
}
