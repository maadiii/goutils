package encryption

import (
	"crypto/rand"
	"io"
	"math/big"
)

type random struct {
	reader io.Reader
}

func NewRandom() *random {
	return &random{reader: rand.Reader}
}

const (
	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func (e *random) String(length int) (string, error) {
	token := make([]byte, length)

	for i := range token {
		num, err := rand.Int(e.reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		token[i] = charset[num.Int64()]
	}

	return string(token), nil
}
