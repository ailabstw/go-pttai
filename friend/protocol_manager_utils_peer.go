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

package friend

import (
	"math/rand"

	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

func (pm *ProtocolManager) LoadPeers() error {
	log.Debug("LoadPeers: start", "entity", pm.Entity().GetID())
	userNodeID, err := pm.GetUserNodeID()
	if err != nil {
		return err
	}

	opKey, err := pm.GetOldestOpKey(false)
	if err != nil {
		return err
	}

	ptt := pm.Ptt()
	ptt.AddDial(userNodeID, opKey.Hash, pkgservice.PeerTypeImportant, true)

	return nil
}

func (pm *ProtocolManager) GetUserNodeID() (*discover.NodeID, error) {

	f := pm.Entity().(*Friend)

	if f.Profile == nil {
		return nil, ErrInvalidFriend
	}

	profilePM := f.Profile.PM().(*account.ProtocolManager)

	nodeList, err := profilePM.GetUserNodeList(nil, 0, pttdb.ListOrderNext, false)
	log.Debug("friend.GetUserNodeID: after get UserNodeList", "nodeList", nodeList, "e", err)
	if err != nil {
		return nil, err
	}

	if len(nodeList) == 0 {
		return nil, types.ErrInvalidID
	}

	randInt := rand.Intn(len(nodeList))

	theNode := nodeList[randInt]

	return theNode.NodeID, nil
}
