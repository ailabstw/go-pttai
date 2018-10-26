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

package bip32

import (
	"testing"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
)

const ()

var (
	tDefaultKey, _      = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	tDefaultKeyBytes    = crypto.FromECDSA(tDefaultKey)
	tDefaultPubKey      = &tDefaultKey.PublicKey
	tDefaultPubKeyBytes = crypto.FromECDSAPub(tDefaultPubKey)

	tDefaultSalt = &types.Salt{
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
		48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
		48, 49,
	}

	tDefaultExtendedKey = &ExtendedKey{
		key:       tDefaultKeyBytes,
		chainCode: tDefaultSalt[:],
		isPrivate: true,
	}

	tDefaultExtendedPubKey = &ExtendedKey{
		key:       tDefaultPubKeyBytes,
		chainCode: tDefaultSalt[:],
		isPrivate: false,
	}

	tChildKeyBytes = []byte{
		231, 31, 91, 193, 25, 175, 250, 235, 33, 178,
		251, 136, 37, 64, 12, 223, 163, 183, 136, 155,
		52, 30, 208, 35, 128, 28, 186, 211, 54, 70,
		81, 106,
	}

	tChildChainCode = []byte{
		154, 80, 160, 103, 137, 127, 89, 129, 75, 31,
		121, 68, 71, 254, 100, 138, 250, 225, 181, 194,
		113, 194, 239, 214, 78, 75, 157, 52, 163, 241,
		228, 212,
	}

	tChildExtendedKey = &ExtendedKey{
		key:       tChildKeyBytes,
		chainCode: tChildChainCode,
		childNum:  1,
		isPrivate: true,
	}

	tChildKey, _         = crypto.ToECDSA(tChildKeyBytes)
	tChildPubKey         = &tChildKey.PublicKey
	tChildKeyPubKeyBytes = crypto.FromECDSAPub(tChildPubKey)

	tChildExtendedPubKey = &ExtendedKey{
		key:       tChildKeyPubKeyBytes,
		chainCode: tChildChainCode,
		childNum:  1,
		isPrivate: false,
	}

	tDefaultSalt2 = &types.Salt{
		50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
		50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
		50, 51, 52, 53, 54, 55, 56, 57, 58, 59,
		50, 51,
	}

	tDefaultExtendedKey2 = &ExtendedKey{
		key:       tDefaultKeyBytes,
		chainCode: tDefaultSalt2[:],
		isPrivate: true,
	}

	tDefaultExtendedPubKey2 = &ExtendedKey{
		key:       tDefaultPubKeyBytes,
		chainCode: tDefaultSalt2[:],
		isPrivate: false,
	}

	tChildKeyBytes2 = []byte{
		208, 45, 79, 200, 18, 12, 235, 87, 183, 137,
		9, 222, 2, 207, 243, 61, 102, 219, 220, 84,
		188, 65, 165, 42, 42, 181, 219, 217, 120, 73,
		39, 80,
	}

	tChildChainCode2 = []byte{
		84, 166, 204, 56, 224, 30, 45, 9, 212, 193,
		132, 104, 249, 159, 22, 102, 204, 36, 221, 42,
		117, 47, 31, 113, 242, 9, 251, 174, 35, 7,
		88, 31,
	}

	tChildExtendedKey2 = &ExtendedKey{
		key:       tChildKeyBytes2,
		chainCode: tChildChainCode2,
		childNum:  1,
		isPrivate: true,
	}

	tChildKey2, _         = crypto.ToECDSA(tChildKeyBytes2)
	tChildPubKey2         = &tChildKey2.PublicKey
	tChildKeyPubKeyBytes2 = crypto.FromECDSAPub(tChildPubKey2)

	tChildExtendedPubKey2 = &ExtendedKey{
		key:       tChildKeyPubKeyBytes2,
		chainCode: tChildChainCode2,
		childNum:  1,
		isPrivate: false,
	}

	tChildKeyBytes3 = []byte{
		239, 251, 161, 157, 199, 156, 206, 225, 97, 252,
		215, 132, 48, 80, 165, 185, 115, 62, 210, 52,
		79, 25, 239, 68, 11, 186, 108, 43, 117, 145,
		124, 58,
	}

	tChildChainCode3 = []byte{
		62, 208, 19, 96, 237, 56, 171, 194, 189, 47,
		47, 205, 102, 154, 123, 91, 225, 48, 200, 116,
		191, 173, 82, 244, 125, 161, 73, 138, 230, 152,
		89, 60,
	}

	tChildExtendedKey3 = &ExtendedKey{
		key:       tChildKeyBytes3,
		chainCode: tChildChainCode3,
		childNum:  2,
		isPrivate: true,
	}

	tChildKey3, _         = crypto.ToECDSA(tChildKeyBytes3)
	tChildPubKey3         = &tChildKey3.PublicKey
	tChildKeyPubKeyBytes3 = crypto.FromECDSAPub(tChildPubKey3)

	tChildExtendedPubKey3 = &ExtendedKey{
		key:       tChildKeyPubKeyBytes3,
		chainCode: tChildChainCode3,
		childNum:  2,
		isPrivate: false,
	}

	origNewSalt func() (*types.Salt, error)
)

func setupTest(t *testing.T) {
	tDefaultExtendedKey.pubKey = nil
	tDefaultExtendedKey2.pubKey = nil

	origNewSalt = types.NewSalt
	types.NewSalt = func() (*types.Salt, error) {
		return &types.Salt{
			48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
			48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
			48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
			48, 49,
		}, nil
	}
}

func teardownTest(t *testing.T) {
	types.NewSalt = origNewSalt
}
