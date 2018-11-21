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

package me

import (
	"github.com/ailabstw/go-pttai/account"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/content"
	"github.com/ailabstw/go-pttai/pttdb"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type PrivateAPI struct {
	b *Backend
}

func NewPrivateAPI(b *Backend) *PrivateAPI {
	return &PrivateAPI{b}
}

/**********
 * Set
 **********/

func (api *PrivateAPI) SetMyName(name []byte) (*account.UserName, error) {
	return api.b.SetMyName(name)
}

func (api *PrivateAPI) SetMyNodeName(nodeID string, name []byte) (*MyNode, error) {
	return api.b.SetMyNodeName([]byte(nodeID), name)
}

func (api *PrivateAPI) SetMyImage(imgStr string) (*account.UserImg, error) {
	return api.b.SetMyImage(imgStr)
}

/**********
 * Revoke
 **********/

func (api *PrivateAPI) Revoke(myKey string) (bool, error) {
	return api.b.Revoke([]byte(myKey))
}

/**********
 * JoinMe
 **********/

func (api *PrivateAPI) ShowMeURL() (*pkgservice.BackendJoinURL, error) {
	return api.b.ShowMeURL()
}

func (api *PrivateAPI) JoinMe(meURL string, myKey string, dummy bool) (*pkgservice.BackendJoinRequest, error) {

	return api.b.JoinMe([]byte(meURL), []byte(myKey))
}

func (api *PrivateAPI) GetJoinKeyInfos(entityID string) ([]*pkgservice.KeyInfo, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetJoinKeys([]byte(entityID))
}

/*
GetMeRequests get the me-requests from me to the others.
*/
func (api *PrivateAPI) GetMeRequests(entityID string) ([]*pkgservice.BackendJoinRequest, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetMeRequests([]byte(entityID))
}

/**********
 * JoinFriend
 **********/

func (api *PrivateAPI) JoinFriend(friendURL string) (*pkgservice.BackendJoinRequest, error) {
	return api.b.JoinFriend([]byte(friendURL))
}

/*
GetFriendRequests get the friend-requests from me to the others.
*/
func (api *PrivateAPI) GetFriendRequests(entityID string) ([]*pkgservice.BackendJoinRequest, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetFriendRequests([]byte(entityID))
}

/**********
 * JoinBoard
 **********/

func (api *PrivateAPI) JoinBoard(friendURL string) (*pkgservice.BackendJoinRequest, error) {
	return api.b.JoinBoard([]byte(friendURL))
}

/*
GetBoardRequests get the friend-requests from me to the others.
*/
func (api *PrivateAPI) GetBoardRequests(entityID string) ([]*pkgservice.BackendJoinRequest, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetBoardRequests([]byte(entityID))
}

/**********
 * Op
 **********/
func (api *PrivateAPI) GetOpKeyInfos(entityID string) ([]*pkgservice.KeyInfo, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetOpKeys([]byte(entityID))
}

func (api *PrivateAPI) RevokeOpKey(entityID string, keyID string, myKey string) (bool, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return false, err
		}
	}
	return api.b.RevokeOpKey([]byte(entityID), []byte(keyID), []byte(myKey))
}

func (api *PrivateAPI) GetOpKeyInfosFromDB(entityID string) ([]*pkgservice.KeyInfo, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetOpKeysFromDB([]byte(entityID))
}

/**********
 * Peer
 **********/

func (api *PrivateAPI) CountPeers(entityID string) (int, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return 0, err
		}
	}
	return api.b.CountPeers([]byte(entityID))
}

func (api *PrivateAPI) GetPeers(entityID string) ([]*pkgservice.BackendPeer, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetPeers([]byte(entityID))
}

/**********
 * My Info
 **********/

func (api *PrivateAPI) GetMyBoard() (*content.BackendGetBoard, error) {
	return api.GetBoard("")
}

func (api *PrivateAPI) GetBoard(entityID string) (*content.BackendGetBoard, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetBoard([]byte(entityID))
}

func (api *PrivateAPI) GetRawMe(entityID string) (*MyInfo, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetRawMe([]byte(entityID))
}

/**********
 * Raft / Node
 **********/

func (api *PrivateAPI) GetRaftStatus(id string) (*RaftStatus, error) {
	return api.b.GetRaftStatus([]byte(id))
}

func (api *PrivateAPI) RemoveNode(nodeID string) (bool, error) {
	return api.b.RemoveNode(nodeID)
}

func (api *PrivateAPI) ForceRemoveNode(nodeID string) (bool, error) {
	return api.b.ForceRemoveNode(nodeID)
}

func (api *PrivateAPI) GetMyNodes() ([]*MyNode, error) {
	return api.GetRawMyNodes("")
}

func (api *PrivateAPI) GetRawMyNodes(entityID string) ([]*MyNode, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetMyNodes([]byte(entityID))
}

func (api *PrivateAPI) GetTotalWeight(entityID string) (uint32, error) {
	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return 0, err
		}
	}
	return api.b.GetTotalWeight([]byte(entityID))
}

func (api *PrivateAPI) RequestRaftLead() (bool, error) {
	return api.b.RequestRaftLead()
}

/**********
 * MeOplog
 **********/

func (api *PrivateAPI) GetMeOplogList(logID string, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {
	return api.GetRawMeOplogList("", logID, limit, listOrder)
}

func (api *PrivateAPI) GetRawMeOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {

	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}

	return api.b.GetMeOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMeOplogMasterList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {

	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetPendingMeOplogMasterList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingMeOplogInternalList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*MeOplog, error) {

	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}

	return api.b.GetPendingMeOplogInternalList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetMeOplogMerkleNodeList(entityID string, level uint8, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.BackendMerkleNode, error) {

	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}

	return api.b.GetMeOplogMerkleNodeList([]byte(entityID), pkgservice.MerkleTreeLevel(level), startKey, limit, listOrder)
}

/**********
 * MasterOplog
 **********/

func (api *PrivateAPI) GetMyMasterOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*MasterOplog, error) {

	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetMyMasterOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

/**********
 * OpKeyOplog
 **********/

func (api *PrivateAPI) GetOpKeyOplogList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {

	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}

	return api.b.GetOpKeyOplogList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingOpKeyOplogMasterList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {

	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}
	return api.b.GetPendingOpKeyOplogMasterList([]byte(entityID), []byte(logID), limit, listOrder)
}

func (api *PrivateAPI) GetPendingOpKeyOplogInternalList(entityID string, logID string, limit int, listOrder pttdb.ListOrder) ([]*pkgservice.OpKeyOplog, error) {

	var err error
	if len(entityID) == 0 {
		entityID, err = api.b.GetMyIDStr()
		if err != nil {
			return nil, err
		}
	}

	return api.b.GetPendingOpKeyOplogInternalList([]byte(entityID), []byte(logID), limit, listOrder)
}

/**********
 * Key
 **********/

func (api *PrivateAPI) ShowMyKey() (*types.PttID, error) {
	return api.b.ShowValidateKey()
}

func (api *PrivateAPI) ValidateMyKey(key string) (bool, error) {
	return api.b.ValidateValidateKey([]byte(key))
}

func (api *PrivateAPI) ShowMyMasterKey() ([]byte, error) {
	return api.b.ShowMyMasterKey()
}

func (api *PrivateAPI) ValidateMyMasterKey(masterKeyBytes []byte) (bool, error) {
	return api.b.ValidateMyMasterKey(masterKeyBytes)
}

func (api *PrivateAPI) ShowMyNodeKey() ([]byte, error) {
	return api.b.ShowMyNodeKey()
}

func (api *PrivateAPI) ValidateMyNodeKey(nodeKeyBytes []byte) (bool, error) {
	return api.b.ValidateMyNodeKey(nodeKeyBytes)
}

func (api *PrivateAPI) ShowMySignKey() (*pkgservice.KeyInfo, error) {
	return api.b.ShowMySignKey()
}

func (api *PrivateAPI) RefreshMySignKey() (*pkgservice.KeyInfo, error) {
	return api.b.RefreshMySignKey()
}

func (api *PrivateAPI) ShowMyNodeSignKey() (*pkgservice.KeyInfo, error) {
	return api.b.ShowMyNodeSignKey()
}

func (api *PrivateAPI) RefreshMyNodeSignKey() (*pkgservice.KeyInfo, error) {
	return api.b.RefreshMyNodeSignKey()
}

/**********
 * Misc
 **********/

func (api *PublicAPI) GetMeList() ([]*BackendMyInfo, error) {
	return api.b.GetMeList()
}

/**********
 * public
 **********/

type PublicAPI struct {
	b *Backend
}

func NewPublicAPI(b *Backend) *PublicAPI {
	return &PublicAPI{b}
}

func (api *PublicAPI) Get() (*BackendMyInfo, error) {
	return api.b.Get()
}

func (api *PublicAPI) ShowURL() (*pkgservice.BackendJoinURL, error) {
	return api.b.ShowURL()
}
