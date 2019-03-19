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

import "github.com/ailabstw/go-pttai/log"

func (pm *BaseProtocolManager) SyncOplogInvalidByMerkle(
	myNewNodes []*MerkleNode,
	theirNewNodes []*MerkleNode,

	forceSyncOplogMsg OpType,
	forceSyncOplogAckMsg OpType,

	merkle *Merkle,

	peer *PttPeer,

) error {

	var err error
	log.Debug("SyncOplogInvalidByMerkle: to ForceSyncOplogByMerkle", "myNewNodes", myNewNodes, "merkle", merkle.Name, "peer", peer)
	for _, node := range myNewNodes {
		err = pm.ForceSyncOplogByMerkle(
			node,

			forceSyncOplogMsg,

			merkle,
			peer,
		)
		if err != nil {
			return err
		}
	}

	if len(theirNewNodes) == 0 {
		return nil
	}

	theirNewKeys := make([][]byte, 0, len(theirNewNodes))
	var theirNewKey []byte
	for _, node := range theirNewNodes {
		theirNewKey = node.ToKey(merkle)
		theirNewKeys = append(theirNewKeys, theirNewKey)
	}

	log.Debug("SyncOplogInvalidByMerkle: to ForceSyncOplogByMerkleAck", "theirNewkeys", theirNewKeys, "merkle", merkle.Name, "peer", peer)

	err = pm.ForceSyncOplogByMerkleAck(
		theirNewKeys,

		forceSyncOplogAckMsg,

		merkle,
		peer,
	)
	if err != nil {
		return err
	}

	return nil
}
