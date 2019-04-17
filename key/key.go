package key

import (
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/ethereum/go-ethereum/crypto"
)

func GenerateKey() (*ecdsa.PrivateKey, error) {
	var key *ecdsa.PrivateKey
	var err error
	for i := 0; i < NGenerateKey; i++ {
		key, err = generateKeyCore()
		if err != nil {
			return nil, err
		}
		if IsValidPrivateKey(key) {
			return key, nil
		}
	}

	return nil, ErrInvalidKey
}

func generateKeyCore() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(crypto.S256(), rand.Reader)
}
