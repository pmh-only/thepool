package main

import (
	"crypto/rand"
	"math/big"
)

func randID(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	out := make([]byte, n)
	max := big.NewInt(int64(len(alphabet)))
	for i := range out {
		k, _ := rand.Int(rand.Reader, max)
		out[i] = alphabet[k.Int64()]
	}
	return string(out)
}
