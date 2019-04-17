package key

import (
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/ethereum/go-ethereum/common"
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

func PubkeyBytesToAddress(pubBytes []byte) common.Address {
	return common.BytesToAddress(crypto.Keccak256(pubBytes[1:])[12:])
}

func PubkeyToHash(p ecdsa.PublicKey) common.Hash {
	pubBytes := crypto.FromECDSAPub(&p)

	return crypto.Keccak256Hash(pubBytes[1:])
}
