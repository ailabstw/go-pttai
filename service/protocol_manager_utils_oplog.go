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
	"reflect"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

func (pm *BaseProtocolManager) IsValidOplog(signInfos []*SignInfo) (*types.PttID, uint32, bool) {
	return pm.isValidOplog(signInfos)
}

func (pm *BaseProtocolManager) defaultIsValidOplog(signInfos []*SignInfo) (*types.PttID, uint32, bool) {

	pm.lockMaster.RLock()
	defer pm.lockMaster.RUnlock()

	lenMaster := len(pm.masters)
	count := 0
	for _, signInfo := range signInfos {
		_, ok := pm.masters[*signInfo.ID]
		if ok {
			count++
		}
	}

	// criteria
	if count < lenMaster {
		return nil, 0, false
	}

	masterOplogID := pm.GetNewestMasterLogID()

	log.Debug("defaultIsValidOplog: to return", "masterLogID", masterOplogID, "count", count)

	return masterOplogID, uint32(count), true
}

/**********
 * BaseProtocolManager
 **********/

func (pm *BaseProtocolManager) SignOplog(oplog *BaseOplog) error {
	myEntity := pm.Ptt().GetMyEntity()

	err := myEntity.Sign(oplog)
	if err != nil {
		return err
	}

	_, err = pm.InternalSign(oplog)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) ForceSignOplog(oplog *BaseOplog) error {
	myEntity := pm.Ptt().GetMyEntity()

	err := myEntity.Sign(oplog)
	if err != nil {
		return err
	}

	if oplog.MasterLogID != nil {
		return nil
	}

	err = myEntity.MasterSign(oplog)
	if err != nil {
		return err
	}

	masterLogID := pm.GetNewestMasterLogID()

	err = oplog.SetMasterLogID(masterLogID, 1)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) GetOplogMerkleNodeList(merkle *Merkle, level MerkleTreeLevel, startKey []byte, limit int, listOrder pttdb.ListOrder) ([]*MerkleNode, error) {

	var err error
	if len(startKey) == 0 {
		startKey, err = merkle.MarshalKey(level, types.ZeroTimestamp)
		if err != nil {
			return nil, err
		}
	}

	iter, err := merkle.GetMerkleIterByKey(startKey, level, listOrder)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	i := 0
	results := make([]*MerkleNode, 0)
	for iter.Next() {
		if limit > 0 && i == limit {
			break
		}

		val := iter.Value()

		eachMerkleNode := &MerkleNode{}
		err := eachMerkleNode.Unmarshal(val)

		if err != nil {
			continue
		}

		results = append(results, eachMerkleNode)

		i++
	}

	return results, nil

}

func (pm *BaseProtocolManager) BroadcastOplog(oplog *BaseOplog, msg OpType, pendingMsg OpType) error {

	// extras
	origExtra := oplog.Extra
	defer func() {
		oplog.Extra = origExtra
	}()
	oplog.Extra = nil

	myService := pm.Entity().Service()
	MyService := pm.Ptt().GetMyService()

	// msg type
	var toSendPeers []*PttPeer
	var op OpType
	peers := pm.peers
	switch {
	case oplog.MasterLogID != nil:
		toSendPeers = peers.PeerList(false)
		op = msg
	case oplog.InternalSigns != nil:
		toSendPeers = peers.MePeerList(false)
		op = pendingMsg
	default:
		if myService == MyService {
			toSendPeers = peers.MePeerList(false)
		} else {
			toSendPeers = peers.ImportantPeerList(false)

		}

		op = pendingMsg
	}

	log.Debug("BroadcastOplog: to SendDataToPeers", "e", pm.Entity().GetID(), "op", op, "toSendPeers", toSendPeers)

	if len(toSendPeers) == 0 {
		// check whether we need to connect to the masters
		if oplog.MasterLogID == nil && oplog.InternalSigns == nil {
			pm.ConnectMaster()
		}

		return nil
	}

	return pm.SendDataToPeers(op, &AddOplog{Oplog: oplog}, toSendPeers)
}

func (pm *BaseProtocolManager) BroadcastOplogs(oplogs []*BaseOplog, msg OpType, pendingMsg OpType) error {

	// extras
	lenOplog := len(oplogs)

	origExtras := make([]interface{}, lenOplog)
	for i, log := range oplogs {
		origExtras[i] = log.Extra
	}
	defer func() {
		for i, log := range oplogs {
			log.Extra = origExtras[i]
		}
	}()

	for _, log := range oplogs {
		log.Extra = nil
	}

	// peer-types
	peers := pm.peers
	peers.RLock()
	mePeerList := peers.MePeerList(true)
	masterPeerList := peers.ImportantPeerList(true)
	allPeerList := peers.PeerList(true)
	peers.RUnlock()

	myService := pm.Entity().Service()
	MyService := pm.Ptt().GetMyService()
	if myService == MyService {
		masterPeerList = mePeerList
	}

	// op-types
	meLogs := make([]*BaseOplog, 0, lenOplog)
	masterLogs := make([]*BaseOplog, 0, lenOplog)
	allLogs := make([]*BaseOplog, 0, lenOplog)

	for _, log := range oplogs {
		switch {
		case log.MasterLogID != nil:
			allLogs = append(allLogs, log)
		case log.InternalSigns != nil:
			meLogs = append(meLogs, log)
		default:
			masterLogs = append(masterLogs, log)
		}
	}

	// send-data-to-peers
	if len(meLogs) != 0 && len(mePeerList) != 0 {
		pm.SendDataToPeers(pendingMsg, &AddOplogs{Oplogs: meLogs}, mePeerList)
	}

	if len(masterLogs) != 0 && len(masterPeerList) != 0 {
		pm.SendDataToPeers(pendingMsg, &AddOplogs{Oplogs: masterLogs}, masterPeerList)
	}

	if len(allLogs) != 0 && len(allPeerList) != 0 {
		pm.SendDataToPeers(msg, &AddOplogs{Oplogs: allLogs}, allPeerList)
	}

	// check whether we need to connect to the masters
	if len(masterLogs) != 0 && len(masterPeerList) == 0 {
		pm.ConnectMaster()
	}

	return nil
}

func (pm *BaseProtocolManager) InternalSign(oplog *BaseOplog) (bool, error) {
	return pm.internalSign(oplog)
}

func (pm *BaseProtocolManager) defaultInternalSign(oplog *BaseOplog) (bool, error) {
	if oplog.MasterLogID != nil {
		return false, nil
	}

	log.Debug("defaultInternalSign: start", "oplog", oplog)

	ptt := pm.Ptt()
	myEntity := ptt.GetMyEntity()
	myID := myEntity.GetID()

	// check
	if !reflect.DeepEqual(myID, oplog.CreatorID) && !pm.isMaster(myID, false) {
		return false, nil
	}

	// already signs master

	log.Debug("defaultInternalSign: to IDInOplogSigns (master-signs)")

	if IDInOplogSigns(myID, oplog.MasterSigns) {
		return false, nil
	}

	// already signs internal
	nodeSignID := myEntity.GetNodeSignID()

	log.Debug("defaultInternalSign: to IDInOplogSigns (internal-signs)")

	if IDInOplogSigns(nodeSignID, oplog.InternalSigns) {
		return false, nil
	}

	// internal-sign
	log.Debug("defaultInternalSign: to InternalSign")
	err := myEntity.InternalSign(oplog)
	if err != nil {
		return false, err
	}

	log.Debug("defaultInternalSign: to IsValidInternalOplog")
	_, _, isValid := myEntity.IsValidInternalOplog(oplog.InternalSigns)
	if !isValid {
		return true, nil
	}

	// master-sign
	log.Debug("defaultInternalSign: to MasterSign")
	err = myEntity.MasterSign(oplog)
	if err != nil {
		return false, err
	}

	log.Debug("defaultInternalSign: to isValidOplog")
	masterLogID, weight, isValid := pm.isValidOplog(oplog.MasterSigns)
	if !isValid {
		return true, nil
	}

	// master-log-id
	log.Debug("defaultInternalSign: to SetMasterLogID")
	err = oplog.SetMasterLogID(masterLogID, weight)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pm *BaseProtocolManager) GetPendingOplogs(setDB func(oplog *BaseOplog), peer *PttPeer, isGetAll bool) ([]*BaseOplog, []*BaseOplog, error) {

	oplog := &BaseOplog{}
	setDB(oplog)

	expireTime, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}
	expireTime.Ts -= int64(ExpireOplogSeconds)

	pendingLogs, err := GetOplogList(oplog, nil, 0, pttdb.ListOrderNext, types.StatusPending, false)
	if err != nil {
		return nil, nil, err
	}

	internalPendingLogs, err := GetOplogList(oplog, nil, 0, pttdb.ListOrderNext, types.StatusInternalPending, false)
	if err != nil {
		return nil, nil, err
	}

	isMyPeer := false
	isMasterPeer := false
	if peer != nil {
		isMyPeer = peer.PeerType == PeerTypeMe
		isMasterPeer = pm.IsMaster(peer.UserID, false)
	}

	lenLogs := len(pendingLogs) + len(internalPendingLogs)
	logs := make([]*BaseOplog, 0, lenLogs)
	failedLogs := make([]*BaseOplog, 0, lenLogs)

	for _, log := range pendingLogs {
		if log.CreateTS.IsLess(expireTime) {
			failedLogs = append(failedLogs, log)
		} else if isMasterPeer || isGetAll {
			logs = append(logs, log)
		}
	}

	for _, log := range internalPendingLogs {
		if log.CreateTS.IsLess(expireTime) {
			failedLogs = append(failedLogs, log)
		} else if isMyPeer || isGetAll {
			logs = append(logs, log)
		}
	}

	return logs, failedLogs, nil
}

func (pm *BaseProtocolManager) IntegrateOplog(
	oplog *BaseOplog,
	isLocked bool,

	merkle *Merkle,
) (bool, error) {
	if !isLocked {
		err := oplog.Lock()
		if err != nil {
			return false, err
		}
		defer oplog.Unlock()
	}

	isToSign, err := oplog.IntegrateExisting(true)
	if err != nil {
		return false, err
	}
	if !isToSign {
		return false, nil
	}

	err = pm.ValidateIntegrateSign(oplog, true)
	if err != nil {
		return false, err
	}

	err = oplog.Save(true, merkle)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pm *BaseProtocolManager) ValidateIntegrateSign(oplog *BaseOplog, isLocked bool) error {
	return pm.validateIntegrateSign(oplog, isLocked)
}

func (pm *BaseProtocolManager) defaultValidateIntegrateSign(oplog *BaseOplog, isLocked bool) (err error) {
	if !isLocked {
		err = oplog.Lock()
		if err != nil {
			return
		}
		defer oplog.Unlock()
	}

	ptt := pm.Ptt()
	myEntity := ptt.GetMyEntity()

	_, weight, isValid := myEntity.IsValidInternalOplog(oplog.InternalSigns)
	if isValid {
		err = myEntity.MasterSign(oplog)
		if err != nil {
			return
		}
	}

	masterLogID, weight, isValid := pm.isValidOplog(oplog.MasterSigns)
	if isValid {
		err = oplog.SetMasterLogID(masterLogID, weight)
		if err != nil {
			return
		}
	}

	return
}

func (pm *BaseProtocolManager) SetOplogIsSync(
	oplog *BaseOplog, isBroadcast bool,
	broadcastLog func(oplog *BaseOplog) error,
) (bool, error) {

	oplog.IsSync = true
	isNewSign, err := pm.InternalSign(oplog)
	if err != nil {
		return false, err
	}

	log.Debug("SetOplogIsSync: to check broadcastLog", "isNewSign", isNewSign, "isBroadcast", isBroadcast, "oplog", oplog.ID)

	if isNewSign && isBroadcast {
		broadcastLog(oplog)
	}

	return isNewSign, nil
}

func (pm *BaseProtocolManager) SetOplogIDIsSync(
	logID *types.PttID,
) (*BaseOplog, error) {
	return nil, types.ErrNotImplemented
}

func (pm *BaseProtocolManager) PostprocessPendingDeleteOplog(oplog *BaseOplog, toBroadcastLogs []*BaseOplog) []*BaseOplog {

	_, err := pm.InternalSign(oplog)
	if err != nil {
		return toBroadcastLogs
	}

	return append(toBroadcastLogs, oplog)
}

func (pm *BaseProtocolManager) CleanOplog(oplog *BaseOplog, merkle *Merkle) {
	db := oplog.GetDB().DB()
	var key []byte
	var val []byte

	iter, err := GetOplogIterWithOplog(oplog, nil, pttdb.ListOrderNext, types.StatusAlive, false)
	if err != nil {
		return
	}
	defer iter.Release()

	for iter.Next() {
		key = iter.Key()
		val = iter.Value()
		err = oplog.Unmarshal(val)
		if err != nil {
			db.Delete(key)
			continue
		}
		oplog.Delete(true)
	}

	iter, err = GetOplogIterWithOplog(oplog, nil, pttdb.ListOrderNext, types.StatusPending, false)
	if err != nil {
		return
	}
	defer iter.Release()

	for iter.Next() {
		key = iter.Key()
		val = iter.Value()
		err = oplog.Unmarshal(val)
		if err != nil {
			db.Delete(key)
			continue
		}
		oplog.Delete(true)
	}

	iter, err = GetOplogIterWithOplog(oplog, nil, pttdb.ListOrderNext, types.StatusInternalPending, false)
	if err != nil {
		return
	}
	defer iter.Release()

	for iter.Next() {
		key = iter.Key()
		val = iter.Value()
		err = oplog.Unmarshal(val)
		if err != nil {
			db.Delete(key)
			continue
		}
		oplog.Delete(true)
	}

	// merkle
	if merkle == nil {
		return
	}

	merkle.Clean()

	return
}
