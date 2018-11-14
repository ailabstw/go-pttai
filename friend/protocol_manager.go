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
	"github.com/ailabstw/go-pttai/common/types"
	pkgservice "github.com/ailabstw/go-pttai/service"
)

type ProtocolManager struct {
	*pkgservice.BaseProtocolManager

	// db
	dbFriendLock      *types.LockMap
	friendOplogMerkle *pkgservice.Merkle
}

func NewProtocolManager(f *Friend, ptt pkgservice.Ptt) (*ProtocolManager, error) {
	dbFriendLock, err := types.NewLockMap(pkgservice.SleepTimeLock)
	if err != nil {
		return nil, err
	}

	friendOplogMerkle, err := pkgservice.NewMerkle(DBFriendOplogPrefix, DBFriendMerkleOplogPrefix, f.ID, dbFriend)
	if err != nil {
		return nil, err
	}
	pm := &ProtocolManager{
		dbFriendLock:      dbFriendLock,
		friendOplogMerkle: friendOplogMerkle,
	}
	b, err := pkgservice.NewBaseProtocolManager(
		ptt, RenewOpKeySeconds, ExpireOpKeySeconds, MaxSyncRandomSeconds, MinSyncRandomSeconds, MaxMasters,
		nil, nil, nil, pm.SetFriendDB,
		nil, nil, nil, nil, nil, nil, nil,
		pm.SyncFriendOplog,  // postsyncMemberOplog
		nil,                 // leave
		pm.DeleteFriend,     // theDelete
		pm.postdeleteFriend, // postdelete
		f, dbFriend)
	if err != nil {
		return nil, err
	}
	pm.BaseProtocolManager = b

	return pm, nil
}
