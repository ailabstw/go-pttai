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

type BoolDoubleArray [][]Bool

func (bdary BoolDoubleArray) MarshalJSON() ([]byte, error) {
	bary2 := [][]Bool(bdary)

	n := len(bary2)
	if n == 0 {
		return nil, nil
	}

	m := len(bary2[0])
	if m == 0 {
		return nil, nil
	}

	out := make([]byte, m*n+1)

	out[0] = uint8(m + 48) // XXX hack for the BoolDoubleArry in that m can not be more than 10 (We just need 2 for our setting for now, m can be extended to 62 based on this setup.)

	idx := 1
	for _, bary := range bary2 {
		for _, b := range bary {
			if b {
				out[idx] = 49
			} else {
				out[idx] = 48
			}

			idx++
		}
	}

	return out, nil
}

func (bdary *BoolDoubleArray) UnmarshalJSON(theBytes []byte) error {
	lenTheBytes := len(theBytes)
	if lenTheBytes < 1 {
		return ErrNotBoolDAry
	}

	m := int(theBytes[0]) - 48
	if m < 0 {
		return ErrNotBoolDAry
	}
	if m > 10 {
		return ErrNotBoolDAry
	}

	n := (lenTheBytes - 1) / m

	if n == 0 || m == 0 {
		return nil
	}

	idx := 1
	pBytes := theBytes[idx:]

	newBDAry := make([][]Bool, n)
	theBool := false
	for i := 0; i < n; i++ {
		newBDAry[i] = make([]Bool, m)
		for j := 0; j < m; j++ {
			if pBytes[0] == 48 {
				theBool = false
			} else {
				theBool = true
			}
			newBDAry[i][j] = Bool(theBool)

			pBytes = pBytes[1:]
		}
	}

	*bdary = BoolDoubleArray(newBDAry)

	return nil
}
