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

package service

import (
	"bytes"
	"reflect"
	"sort"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
)

func ValidateMerkleTree(
	myNodes []*MerkleNode,
	theirNodes []*MerkleNode,
	ts types.Timestamp,

	pm ProtocolManager,
	merkle *Merkle,

) (types.Timestamp, bool) {
	myNodes = validateMerkleTreeTrimNodes(myNodes, ts, pm, merkle)
	theirNodes = validateMerkleTreeTrimNodes(theirNodes, ts, pm, merkle)

	lenMyNodes := len(myNodes)
	lenTheirNodes := len(theirNodes)

	var diffTS types.Timestamp

	i := 0
	for pMyNode, pTheirNode := myNodes, theirNodes; i < lenMyNodes && i < lenTheirNodes; i, pMyNode, pTheirNode = i+1, pMyNode[1:], pTheirNode[1:] {
		if !reflect.DeepEqual(pMyNode[0].Addr, pTheirNode[0].Addr) {
			log.Error("ValidateMerkleTree: invalid", "i", i, "len", lenMyNodes, "myNode", pMyNode[0], "theirNode", pTheirNode[0], "name", merkle.Name)

			diffTS = pMyNode[0].UpdateTS
			if pTheirNode[0].UpdateTS.IsLess(ts) {
				diffTS = pTheirNode[0].UpdateTS
			}
			return diffTS, false
		}
	}

	if i < lenMyNodes {
		return myNodes[i].UpdateTS, false
	}

	if i < lenTheirNodes {
		return theirNodes[i].UpdateTS, false
	}

	return types.ZeroTimestamp, true
}

func validateMerkleTreeTrimNodes(
	nodes []*MerkleNode,
	ts types.Timestamp,

	pm ProtocolManager,
	merkle *Merkle,

) []*MerkleNode {

	nNodes := len(nodes)
	idx := sort.Search(nNodes, func(i int) bool {
		return ts.IsLessEqual(nodes[i].UpdateTS)
	})

	return nodes[:idx]
}

func DiffMerkleTree(
	myNodes []*MerkleNode,
	theirNodes []*MerkleNode,

	ts types.Timestamp,

	pm ProtocolManager,
	merkle *Merkle,

) ([]*MerkleNode, []*MerkleNode, error) {

	if !ts.IsEqual(types.ZeroTimestamp) {
		myNodes = validateMerkleTreeTrimNodes(myNodes, ts, pm, merkle)
		theirNodes = validateMerkleTreeTrimNodes(theirNodes, ts, pm, merkle)
	}

	lenMyNodes := len(myNodes)
	lenTheirNodes := len(theirNodes)

	pMyNodes := myNodes
	pTheirNodes := theirNodes

	var myNode *MerkleNode
	var theirNode *MerkleNode

	myNewNodes := make([]*MerkleNode, 0, lenTheirNodes)
	theirNewNodes := make([]*MerkleNode, 0, lenMyNodes)

	log.Debug("DiffMerkleTree: to for-loop", "myNodes", myNodes, "theirNodes", theirNodes, "ts", ts, "merkle", merkle.Name)

	for len(pMyNodes) > 0 && len(pTheirNodes) > 0 {
		myNode = pMyNodes[0]
		theirNode = pTheirNodes[0]

		switch {
		case myNode.UpdateTS.IsLess(theirNode.UpdateTS):
			log.Error("DiffMerkleTree: myNode.TS", "me", myNode.UpdateTS, "me.level", myNode.Level, "they", theirNode.UpdateTS, "they.level", theirNode.Level, "merkle", merkle.Name)
			theirNewNodes = append(theirNewNodes, myNode)
			pMyNodes = pMyNodes[1:]
		case theirNode.UpdateTS.IsLess(myNode.UpdateTS):
			log.Error("DiffMerkleTree: theirNode.TS", "me", myNode.UpdateTS, "me.level", myNode.Level, "they", theirNode.UpdateTS, "they.level", theirNode.Level, "merkle", merkle.Name)
			myNewNodes = append(myNewNodes, theirNode)
			pTheirNodes = pTheirNodes[1:]
		case myNode.Level > theirNode.Level:
			log.Error("DiffMerkleTree: myNode.Level", "ts", myNode.UpdateTS, "me", myNode.Level, "they", theirNode.Level, "merkle", merkle.Name)
			theirNewNodes = append(theirNewNodes, myNode)
			pMyNodes = pMyNodes[1:]
		case myNode.Level < theirNode.Level:
			log.Error("DiffMerkleTree: theirNode.Level", "ts", myNode.UpdateTS, "me", myNode.Level, "they", theirNode.Level, "merkle", merkle.Name)
			myNewNodes = append(myNewNodes, theirNode)
			pTheirNodes = pTheirNodes[1:]
		case myNode.NChildren > theirNode.NChildren:
			log.Error("DiffMerkleTree: myNode.NChildren", "ts", myNode.UpdateTS, "level", myNode.Level, "me", myNode.NChildren, "they", theirNode.NChildren, "merkle", merkle.Name)
			theirNewNodes = append(theirNewNodes, myNode)
			pMyNodes = pMyNodes[1:]
			pTheirNodes = pTheirNodes[1:]
		case myNode.NChildren < theirNode.NChildren:
			log.Error("DiffMerkleTree: theirNode.NChildren", "ts", myNode.UpdateTS, "level", myNode.Level, "me", myNode.NChildren, "they", theirNode.NChildren, "merkle", merkle.Name)
			myNewNodes = append(myNewNodes, theirNode)
			pMyNodes = pMyNodes[1:]
			pTheirNodes = pTheirNodes[1:]
		default:
			if !reflect.DeepEqual(myNode.Addr, theirNode.Addr) {
				log.Error("DiffMerkleTree: Addr", "ts", myNode.UpdateTS, "level", myNode.Level, "me", myNode, "they", theirNode, "merkle", merkle.Name)
				myNewNodes = append(myNewNodes, theirNode)
				theirNewNodes = append(theirNewNodes, myNode)
			}

			pMyNodes = pMyNodes[1:]
			pTheirNodes = pTheirNodes[1:]
		}
	}

	if len(pMyNodes) > 0 {
		log.Error("DiffMerkleTree: myNodes", "ts", pMyNodes[0].UpdateTS, "level", pMyNodes[0].Level, "merkle", merkle.Name)
		theirNewNodes = append(theirNewNodes, pMyNodes...)
	}

	if len(pTheirNodes) > 0 {
		log.Error("DiffMerkleTree: theirNodes", "ts", pTheirNodes[0].UpdateTS, "level", pTheirNodes[0].Level, "merkle", merkle.Name)
		myNewNodes = append(myNewNodes, pTheirNodes...)
	}

	return myNewNodes, theirNewNodes, nil
}

/*
Return: myNewKeys: new keys from from their nodes, theirNewKeys: new keys from my nodes
*/
func MergeKeysInMerkleNodes(myNodes []*MerkleNode, theirNodes []*MerkleNode) ([][]byte, [][]byte, error) {
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

func DiffMerkleKeys(
	myKeys [][]byte,
	theirKeys [][]byte,
) ([][]byte, [][]byte, error) {

	lenMyKeys := len(myKeys)
	lenTheirKeys := len(theirKeys)

	pMyKeys := myKeys
	pTheirKeys := theirKeys

	var myKey []byte
	var theirKey []byte

	myNewKeys := make([][]byte, 0, lenTheirKeys)
	theirNewKeys := make([][]byte, 0, lenMyKeys)

	cmp := 0
	for len(pMyKeys) > 0 && len(pTheirKeys) > 0 {
		myKey = pMyKeys[0]
		theirKey = pTheirKeys[0]

		cmp = bytes.Compare(myKey, theirKey)
		switch {
		case cmp < 0:
			theirNewKeys = append(theirNewKeys, myKey)
			pMyKeys = pMyKeys[1:]
		case cmp > 0:
			myNewKeys = append(myNewKeys, theirKey)
			pTheirKeys = pTheirKeys[1:]
		default:
			pMyKeys = pMyKeys[1:]
			pTheirKeys = pTheirKeys[1:]
		}
	}

	if len(pMyKeys) > 0 {
		theirNewKeys = append(theirNewKeys, pMyKeys...)
	}

	if len(pTheirKeys) > 0 {
		myNewKeys = append(myNewKeys, pTheirKeys...)
	}

	return myNewKeys, theirNewKeys, nil
}

func getKeysFromMerkleKeys(merkle *Merkle, merkleKeys [][]byte) ([][]byte, error) {
	keys := make([][]byte, 0, len(merkleKeys))

	var err error
	var merkleNode *MerkleNode
	for _, key := range merkleKeys {
		merkleNode, err = merkle.GetNodeByKey(key)
		if err != nil {
			continue
		}

		if merkleNode.Level != MerkleTreeLevelNow {
			log.Warn("getKeysFromMerkleKeys: wrong level", "key", key, "level", merkleNode.Level, "ts", merkleNode.UpdateTS, "merkle", merkle.Name)
			continue
		}

		keys = append(keys, merkleNode.Key)
	}

	return keys, nil
}

func GetMerkleName(merkle *Merkle, pm ProtocolManager) string {
	if merkle != nil {
		return merkle.Name
	}

	return pm.Entity().IDString()
}
