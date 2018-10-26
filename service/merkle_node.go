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
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
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

	return common.Concat([][]byte{[]byte{uint8(m.Level)}, m.Addr, tsBytes, childrenBytes, m.Key})
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
	key := common.CloneBytes(b[offset:])

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

/*
Return: myNewKeys: new keys from from their nodes, theirNewKeys: new keys from my nodes
*/
func MergeMerkleNodeKeys(myNodes []*MerkleNode, theirNodes []*MerkleNode) ([][]byte, [][]byte, error) {
	// XXX TODO: refactor. Currently the 4 conditions are enumerated.
	lenMyNodes := len(myNodes)
	lenTheirNodes := len(theirNodes)

	/*
		for i, myNode := range myNodes {
			log.Debug("MergeMerkleNodeKeys: to for-loop", "idx", fmt.Sprintf("(%d/%d)", i, lenMyNodes), "myNode", myNode)
		}

		for i, theirNode := range theirNodes {
			log.Debug("MergeMerkleNodeKeys: to for-loop", "idx", fmt.Sprintf("(%d/%d)", i, lenTheirNodes), "theirNode", theirNode)
		}
	*/

	myNewKeys := make([][]byte, 0, lenTheirNodes)
	theirNewKeys := make([][]byte, 0, lenMyNodes)

	if lenMyNodes == 0 && lenTheirNodes == 0 {
		return myNewKeys, theirNewKeys, nil
	}

	var myKey []byte = nil
	var theirKey []byte = nil
	if lenTheirNodes == 0 {
		for pMyNodes, myIdx := myNodes, 0; myIdx < lenMyNodes; pMyNodes, myIdx = pMyNodes[1:], myIdx+1 {
			myKey = pMyNodes[0].Key
			theirNewKeys = append(theirNewKeys, myKey)
		}
		return myNewKeys, theirNewKeys, nil
	}

	if lenMyNodes == 0 {
		for pTheirNodes, theirIdx := theirNodes, 0; theirIdx < lenTheirNodes; pTheirNodes, theirIdx = pTheirNodes[1:], theirIdx+1 {
			theirKey = pTheirNodes[0].Key
			myNewKeys = append(myNewKeys, theirKey)
		}
		return myNewKeys, theirNewKeys, nil
	}

	pMyNodes := myNodes
	myIdx := 0
	myKey = pMyNodes[0].Key

	pTheirNodes := theirNodes
	theirIdx := 0
	theirKey = pTheirNodes[0].Key

	for myIdx < lenMyNodes && theirIdx < lenTheirNodes {
		cmp := bytes.Compare(myKey, theirKey)
		//log.Debug("MergeMerkleNodeKeys: after cmp", "idx", fmt.Sprintf("(%d/%d/%d/%d)", myIdx, lenMyNodes, theirIdx, lenTheirNodes), "myKey", myKey, "theirKey", theirKey, "cmp", cmp)
		if cmp < 0 { // myKey < theirKey
			theirNewKeys = append(theirNewKeys, myKey)
			myIdx++
			if myIdx == lenMyNodes {
				break
			}
			pMyNodes = pMyNodes[1:]
			myKey = pMyNodes[0].Key
		} else if cmp > 0 { // myKey > theirKey
			myNewKeys = append(myNewKeys, theirKey)
			theirIdx++
			if theirIdx == lenTheirNodes {
				break
			}
			pTheirNodes = pTheirNodes[1:]
			theirKey = pTheirNodes[0].Key
		} else { // myKey == theirKey
			myIdx++
			if myIdx < lenMyNodes {
				pMyNodes = pMyNodes[1:]
				myKey = pMyNodes[0].Key
			}

			theirIdx++
			if theirIdx < lenTheirNodes {
				pTheirNodes = pTheirNodes[1:]
				theirKey = pTheirNodes[0].Key
			}
		}
	}

	//log.Debug("MergeMerkleNodeKeys: after for-loop", "myIdx", myIdx, "lenMyNodes", lenMyNodes, "theirNewKeys", theirNewKeys, "theirIdx", theirIdx, "lenTheirNodes", lenTheirNodes, "myNewKeys", myNewKeys)

	for myIdx < lenMyNodes {
		theirNewKeys = append(theirNewKeys, myKey)
		myIdx++
		if myIdx == lenMyNodes {
			break
		}
		pMyNodes = pMyNodes[1:]
		myKey = pMyNodes[0].Key
	}

	for theirIdx < lenTheirNodes {
		myNewKeys = append(myNewKeys, theirKey)
		theirIdx++
		if theirIdx == lenTheirNodes {
			break
		}
		pTheirNodes = pTheirNodes[1:]
		theirKey = pTheirNodes[0].Key
	}

	//log.Debug("MergeMerkleNodeKeys: after for-loop", "myIdx", myIdx, "lenMyNodes", lenMyNodes, "theirNewKeys", theirNewKeys, "theirIdx", theirIdx, "lenTheirNodes", lenTheirNodes, "myNewKeys", myNewKeys)

	return myNewKeys, theirNewKeys, nil
}
