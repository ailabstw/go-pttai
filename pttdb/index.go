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

package pttdb

import (
	"encoding/json"

	"github.com/ailabstw/go-pttai/common/types"
)

/*
Index stores the information of index for ts-based records.

The 0th key is the key for the record.

The other keys are the secondary index-keys
*/
type Index struct {
	Keys     [][]byte        `json:"K"`
	UpdateTS types.Timestamp `json:"UT"`
}

func IndexGetKey(theBytes []byte) ([]byte, error) {
	i := &Index{}
	err := i.Unmarshal(theBytes)
	if err != nil {
		return nil, err
	}

	return i.Keys[0], nil
}

func IndexGetKeys(theBytes []byte) ([][]byte, error) {
	i := &Index{}
	err := i.Unmarshal(theBytes)
	if err != nil {
		return nil, err
	}

	return i.Keys, nil
}

func (i *Index) Marshal() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Index) Unmarshal(data []byte) error {
	return json.Unmarshal(data, i)
}

type IndexWithStatus struct {
	Keys     [][]byte        `json:"K"`
	UpdateTS types.Timestamp `json:"UT"`
	Status   types.Status    `json:"S"`
}

func (i *IndexWithStatus) Marshal() ([]byte, error) {
	return json.Marshal(i)
}

func (i *IndexWithStatus) Unmarshal(data []byte) error {
	return json.Unmarshal(data, i)
}
