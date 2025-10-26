package encryption

import (
	"crypto/rand"
	"math/big"
)

type random struct{}

func NewRandom() *random {
	return new(random)
}

const (
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func (e *random) String(length int) (string, error) {
	token := make([]byte, length)

	for i := range token {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		token[i] = charset[num.Int64()]
	}

	return string(token), nil
}
