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
	"reflect"
	"sort"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

func (pm *BaseProtocolManager) IsValidOplog(signInfos []*SignInfo) (*types.PttID, uint32, bool) {
	return nil, 0, false
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
		toSendPeers = peers.ImportantPeerList(false)
		op = pendingMsg
	}

	if len(toSendPeers) == 0 {
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

	return nil
}

func (pm *BaseProtocolManager) InternalSign(oplog *BaseOplog) (bool, error) {
	if oplog.MasterLogID != nil {
		return false, nil
	}

	ptt := pm.Ptt()
	myEntity := ptt.GetMyEntity()
	myID := myEntity.GetID()

	// check
	if !reflect.DeepEqual(myID, oplog.CreatorID) && !pm.isMaster(myID) {
		return false, nil
	}

	masterSigns := oplog.MasterSigns
	lenMasterSigns := len(masterSigns)

	// already signs master
	if lenMasterSigns > 0 {
		idx := sort.Search(len(masterSigns), func(i int) bool {
			return bytes.Compare(masterSigns[i].ID[:], myID[:]) >= 0
		})
		if idx >= 0 && idx < lenMasterSigns && reflect.DeepEqual(masterSigns[idx].ID, myID) {
			return false, nil
		}
	}

	nodeSignID := myEntity.GetNodeSignID()

	internalSigns := oplog.InternalSigns
	lenInternalSigns := len(internalSigns)

	// already signs internal
	if lenInternalSigns > 0 {
		idx := sort.Search(len(internalSigns), func(i int) bool {
			return bytes.Compare(internalSigns[i].ID[:], nodeSignID[:]) >= 0
		})
		if idx >= 0 && idx < lenInternalSigns && reflect.DeepEqual(internalSigns[idx].ID, nodeSignID) {
			return false, nil
		}
	}

	// internal-sign
	err := myEntity.InternalSign(oplog)
	if err != nil {
		return false, err
	}

	_, weight, isValid := myEntity.IsValidInternalOplog(oplog.InternalSigns)
	if !isValid {
		return true, nil
	}

	// master-sign
	err = myEntity.MasterSign(oplog)
	if err != nil {
		return false, err
	}

	masterLogID, weight, isValid := pm.isValidOplog(oplog.MasterSigns)
	if !isValid {
		return true, nil
	}

	// master-log-id
	err = oplog.SetMasterLogID(masterLogID, weight)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pm *BaseProtocolManager) GetPendingOplogs(setDB func(oplog *BaseOplog)) ([]*BaseOplog, []*BaseOplog, error) {

	oplog := &BaseOplog{}
	setDB(oplog)

	expireTime, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}
	expireTime.Ts -= ExpireOplogSeconds

	pendingLogs, err := GetOplogList(oplog, nil, 0, pttdb.ListOrderNext, types.StatusPending, false)
	if err != nil {
		return nil, nil, err
	}

	internalPendingLogs, err := GetOplogList(oplog, nil, 0, pttdb.ListOrderNext, types.StatusInternalPending, false)
	if err != nil {
		return nil, nil, err
	}

	lenLogs := len(pendingLogs) + len(internalPendingLogs)
	logs := make([]*BaseOplog, 0, lenLogs)
	failedLogs := make([]*BaseOplog, 0, lenLogs)

	for _, log := range pendingLogs {
		if log.CreateTS.IsLess(expireTime) {
			failedLogs = append(failedLogs, log)
		} else {
			logs = append(logs, log)
		}
	}

	for _, log := range internalPendingLogs {
		if log.CreateTS.IsLess(expireTime) {
			failedLogs = append(failedLogs, log)
		} else {
			logs = append(logs, log)
		}
	}

	return logs, failedLogs, nil
}

func (pm *BaseProtocolManager) IntegrateOplog(oplog *BaseOplog, isLocked bool) (bool, error) {
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

	err = pm.validateIntegrateSign(oplog, true)
	if err != nil {
		return false, err
	}

	err = oplog.Save(true)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pm *BaseProtocolManager) validateIntegrateSign(oplog *BaseOplog, isLocked bool) error {
	var err error
	if !isLocked {
		err = oplog.Lock()
		if err != nil {
			return err
		}
		defer oplog.Unlock()
	}

	ptt := pm.Ptt()
	myEntity := ptt.GetMyEntity()

	_, weight, isValid := myEntity.IsValidInternalOplog(oplog.InternalSigns)
	if isValid {
		err = myEntity.MasterSign(oplog)
		if err != nil {
			return err
		}
	}

	masterLogID, weight, isValid := pm.isValidOplog(oplog.MasterSigns)
	if isValid {
		err = oplog.SetMasterLogID(masterLogID, weight)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *BaseProtocolManager) RemoveNonSyncOplog(setDB func(oplog *BaseOplog), logID *types.PttID, isRetainValid bool, isLocked bool) (*BaseOplog, error) {

	oplog := &BaseOplog{}
	setDB(oplog)
	oplog.ID = logID

	if !isLocked {
		err := oplog.Lock()
		if err != nil {
			return nil, err
		}
		defer oplog.Unlock()
	}

	err := oplog.Get(logID, true)
	if err != nil {
		return nil, err
	}

	status := oplog.ToStatus()
	if oplog.IsSync && status == types.StatusAlive {
		return nil, nil
	}

	if isRetainValid && status == types.StatusAlive {
		oplog.IsSync = true
		err = oplog.SaveWithIsSync(true)
		return oplog, nil
	}

	err = oplog.Delete(true)

	return nil, err
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
