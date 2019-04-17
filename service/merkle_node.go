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

package service

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"

	pttcommon "github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ethereum/go-ethereum/common"
)

type MerkleNode struct {
	Level     MerkleTreeLevel `json:"L"`
	Addr      []byte          `json:"A"`
	UpdateTS  types.Timestamp `json:"UT"`
	NChildren uint32          `json:"N"`
	Key       []byte          `json:"K"`
}

func (m *MerkleNode) Marshal() ([]byte, error) {
	tsBytes, err := m.UpdateTS.Marshal()
	if err != nil {
		return nil, err
	}

	childrenBytes := make([]byte, SizeMerkleTreeNChildren)
	binary.BigEndian.PutUint32(childrenBytes, m.NChildren)

	return pttcommon.Concat([][]byte{[]byte{uint8(m.Level)}, m.Addr, tsBytes, childrenBytes, m.Key})
}

func (m *MerkleNode) Unmarshal(b []byte) error {
	// level
	offset := 0
	level := MerkleTreeLevel(b[0])

	// addr
	offset += SizeMerkleTreeLevel
	addr := make([]byte, common.AddressLength)
	copy(addr, b[offset:])

	offset += common.AddressLength

	// ts
	ts, err := types.UnmarshalTimestamp(b[offset:(offset + types.SizeTimestamp)])
	if err != nil {
		return err
	}

	// n-children
	offset += types.SizeTimestamp
	nChildren := binary.BigEndian.Uint32(b[offset:])

	offset += SizeMerkleTreeNChildren
	key := common.CopyBytes(b[offset:])

	m.Level = level
	m.Addr = addr
	m.UpdateTS = ts
	m.NChildren = nChildren
	m.Key = key

	return nil
}

func (m *MerkleNode) MarshalJSON() ([]byte, error) {
	theBytes, err := m.Marshal()
	if err != nil {
		return nil, err
	}
	return json.Marshal(theBytes)
}

func (m *MerkleNode) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		b = b[1:(len(b) - 1)]
	}

	d, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return err
	}

	return m.Unmarshal(d)
}

func (m *MerkleNode) ConstructUpdateTSAndLevelByKey(key []byte) error {
	level := MerkleTreeLevel(key[MerkleTreeKeyOffsetLevel])
	updateTS, err := types.UnmarshalTimestamp(key[MerkleTreeKeyOffsetUpdateTS:])
	if err != nil {
		return err
	}

	m.Level = level
	m.UpdateTS = updateTS

	return nil
}

func (m *MerkleNode) ToKey(merkle *Merkle) []byte {
	if m.Level == MerkleTreeLevelNow {
		return m.Key
	}

	ts := m.UpdateTSToTS()

	key, _ := merkle.MarshalKey(m.Level, ts)
	return key
}

func (m *MerkleNode) UpdateTSToTS() types.Timestamp {
	var ts types.Timestamp
	switch m.Level {
	case MerkleTreeLevelNow:
		ts = m.UpdateTS
	case MerkleTreeLevelHR:
		ts, _ = m.UpdateTS.ToHRTimestamp()
	case MerkleTreeLevelDay:
		ts, _ = m.UpdateTS.ToDayTimestamp()
	case MerkleTreeLevelMonth:
		ts, _ = m.UpdateTS.ToMonthTimestamp()
	case MerkleTreeLevelYear:
		ts, _ = m.UpdateTS.ToYearTimestamp()
	}

	return ts
}
