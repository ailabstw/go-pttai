// Copyright 2018 The go-pttai Authors
// This file is part of the go-pttai library.
//
// The go-pttai library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-pttai library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-pttai library. If not, see <http://www.gnu.org/licenses/>.

package crypto

import (
	"crypto/ecdsa"
	"reflect"

	"github.com/ailabstw/go-pttai/log"
	"github.com/btcsuite/btcd/btcec"
	p2pcrypto "github.com/libp2p/go-libp2p-crypto"
)

func PrivateKeyToP2PPrivKey(key *ecdsa.PrivateKey) (p2pcrypto.PrivKey, error) {
	return (*p2pcrypto.Secp256k1PrivateKey)(key), nil
}

func PrivateKeyToP2PKey(key *ecdsa.PrivateKey) (p2pcrypto.PrivKey, p2pcrypto.PubKey, error) {
	privKey, err := PrivateKeyToP2PPrivKey(key)
	if err != nil {
		return nil, nil, err
	}

	pubKey := privKey.GetPublic()
	return privKey, pubKey, nil
}

func PubKeyToP2PPubkey(key *ecdsa.PublicKey) (p2pcrypto.PubKey, error) {
	return (*p2pcrypto.Secp256k1PublicKey)(key), nil
}

/*
IsValidPrivateKey is to ensure that crypto.S256() is the same as btcec.S256()
(p2pcrypto is using btcec.S256, while ecsda is using crypto.S256)
*/
func IsValidPrivateKey(key *ecdsa.PrivateKey) bool {
	marshaledPrivKey := FromECDSA(key)
	marshaledPubKey := FromECDSAPub(&key.PublicKey)
	p2pPrivKey, err := p2pcrypto.UnmarshalSecp256k1PrivateKey(marshaledPrivKey)
	if err != nil {
		log.Warn("IsValidPrivateKey: unable to unmarshal private key", "e", err)
		return false
	}
	p2pPubKey := p2pPrivKey.GetPublic()
	marshaledP2PPubKey, err := p2pPubKey.Raw()
	if err != nil {
		log.Warn("IsValidPrivateKey: unable to p2pPubKey.Raw", "e", err)
		return false
	}
	btcPubKey, err := btcec.ParsePubKey(marshaledP2PPubKey, btcec.S256())
	if err != nil {
		log.Warn("IsValidPrivateKey: unable to btcec.ParsePubKey", "e", err)
		return false
	}
	pubKey := btcPubKey.ToECDSA()
	marshaledPubKey2 := FromECDSAPub(pubKey)
	if !reflect.DeepEqual(marshaledPubKey, marshaledPubKey2) {
		log.Warn("IsValidPrivateKey: marshaledPubKey not the same", "marshaledPubKey", marshaledPubKey, "marshaledPubKey2", marshaledPubKey2)
		return false
	}

	return true
}

func P2PPubKeyBytesToPubKey(theBytes []byte) (*ecdsa.PublicKey, error) {
	btcPubKey, err := btcec.ParsePubKey(theBytes, btcec.S256())
	if err != nil {
		log.Warn("P2PPubKeyBytesToPubKey: unable to btcec.ParsePubKey", "e", err)
		return nil, err
	}
	return btcPubKey.ToECDSA(), nil
}

func P2PPubKeyBytesToPubKeyBytes(theBytes []byte) ([]byte, error) {
	pubKey, err := P2PPubKeyBytesToPubKey(theBytes)
	if err != nil {
		return nil, err
	}
	return FromECDSAPub(pubKey), nil
}
