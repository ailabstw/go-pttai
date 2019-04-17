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

package me

import (
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) IsValidInternalOplog(signInfos []*pkgservice.SignInfo) (*types.PttID, uint32, bool) {
	pm.lockMyNodes.RLock()
	defer pm.lockMyNodes.RUnlock()

	weight := uint32(0)
	var node *MyNode
	for _, signInfo := range signInfos {
		node = pm.MyNodeByNodeSignIDs[*signInfo.ID]
		if node == nil {
			continue
		}
		weight += node.Weight
	}

	masterOplogID := pm.GetNewestMasterLogID()

	isValid := weight >= pm.Quorum()
	if !isValid {
		return nil, 0, false
	}

	return masterOplogID, weight, weight >= pm.Quorum()
}

func (pm *ProtocolManager) Quorum() uint32 {
	return pm.totalWeight/2 + 1
}

func (pm *ProtocolManager) nodeTypeToWeight(nodeType pkgservice.NodeType) uint32 {

	switch nodeType {
	case pkgservice.NodeTypeServer:
		return WeightServer
	case pkgservice.NodeTypeDesktop:
		return WeightDesktop
	case pkgservice.NodeTypeMobile:
		return WeightMobile
	}

	return 0
}
