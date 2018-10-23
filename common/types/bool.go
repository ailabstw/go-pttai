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

package types

type Bool bool

func (b Bool) MarshalJSON() ([]byte, error) {
	if bool(b) {
		return []byte{49}, nil
	}

	return []byte{48}, nil
}

func (b *Bool) UnmarshalJSON(theByte []byte) error {
	if len(theByte) != 1 {
		return ErrNotBool
	}

	*b = Bool(theByte[0] == 49)

	return nil
}
