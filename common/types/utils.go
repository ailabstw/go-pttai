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
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetBoolFromInterface(v interface{}, name string) bool {
	val := reflect.ValueOf(v).FieldByName(name)

	if !val.IsValid() || val.Kind() != reflect.Bool {
		return false
	}

	return val.Bool()
}

func Hash(theBytes ...[]byte) []byte {
	hash := crypto.Keccak256Hash(theBytes...)
	return hash[OffsetHash:(OffsetHash + common.AddressLength)]
}

func Addr(theBytes ...[]byte) []byte {
	hash := crypto.Keccak256Hash(theBytes...)
	return hash[OffsetHash:(OffsetHash + common.AddressLength)]
}

func HashToAddr(theBytes []byte) []byte {
	return theBytes[OffsetHash:(OffsetHash + common.AddressLength)]
}
