// Copyright 2019 The go-pttai Authors
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

package types

import (
	"crypto/rand"

	"github.com/ailabstw/go-pttai/common"
	"github.com/shengdoushi/base58"
)

type Salt [SizeSalt]byte

var NewSalt = func() (*Salt, error) {
	salt := &Salt{}
	rand.Read(salt[:])

	return salt, nil
}

func (s *Salt) MarshalJSON() ([]byte, error) {
	marshaled := []byte(base58.Encode(s[:], myAlphabet))
	return common.Concat([][]byte{quoteBytes, marshaled, quoteBytes})
}

func (s *Salt) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return ErrInvalidID
	}

	if b[0] == '"' { // hack for possibly json-stringify strings
		b = b[1:(len(b) - 1)]
	}

	decodedBytes, err := base58.Decode(string(b), myAlphabet)
	if err != nil {
		return err
	}

	if len(decodedBytes) != SizeSalt {
		return ErrInvalidID
	}

	copy(s[:], decodedBytes)

	return nil
}
