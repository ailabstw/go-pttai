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
	return pm.isValidOplog(signInfos)
}

/**********
 * BaseProtocolManager
 **********/

func (pm *BaseProtocolManager) SignOplog(oplog *Oplog) error {
	ptt := pm.Ptt()
	key := ptt.SignKey()

	err := oplog.Sign(key)
	if err != nil {
		return err
	}

	_, err = pm.InternalSign(oplog)
	if err != nil {
		return err
	}

	return nil
}

func (pm *BaseProtocolManager) GetOplogsFromKeys(setDB func(log *Oplog), keys [][]byte) ([]*Oplog, error) {

	logs := make([]*Oplog, 0, len(keys))
	var log *Oplog
	for _, key := range keys {
		log = &Oplog{}
		setDB(log)
		err := log.Load(key)
		if err != nil {
			continue
		}

		logs = append(logs, log)
	}

	return logs, nil
}

func (pm *BaseProtocolManager) GetOplogList(log *Oplog, startID *types.PttID, limit int, listOrder pttdb.ListOrder, status types.Status, isLocked bool) ([]*Oplog, error) {

	iter, err := GetOplogIter(log.GetDB(), log.GetDBPrefix(), log.GetDBIdxPrefix(), log.GetDBMerklePrefix(), log.GetDBPrefxiID(), startID, log.GetDBLock(), isLocked, status, listOrder)
	if err != nil {
		return nil, err
	}
	defer iter.Release()

	funcIter := pttdb.GetFuncIter(iter, listOrder)

	// for-loop
	var eachLog *Oplog
	logs := make([]*Oplog, 0)
	i := 0
	for funcIter() {
		if limit > 0 && i >= limit {
			break
		}

		v := iter.Value()

		eachLog = &Oplog{}
		err := log.Unmarshal(v)
		if err != nil {
			continue
		}

		logs = append(logs, eachLog)

		i++
	}

	return logs, nil
}

func (pm *BaseProtocolManager) BroadcastOplog(log *Oplog, msg OpType, pendingMsg OpType) error {

	// extras
	origExtra := log.Extra
	defer func() {
		log.Extra = origExtra
	}()
	log.Extra = nil

	// msg type
	var toSendPeers []*PttPeer
	var op OpType
	peers := pm.peers
	switch {
	case log.MasterLogID != nil:
		toSendPeers = peers.PeerList(false)
		op = msg
	case log.InternalSigns != nil:
		toSendPeers = peers.MePeerList(false)
		op = pendingMsg
	default:
		toSendPeers = peers.ImportantPeerList(false)
		op = pendingMsg
	}

	if len(toSendPeers) == 0 {
		return nil
	}

	return pm.SendDataToPeers(op, log, toSendPeers)
}

func (pm *BaseProtocolManager) BroadcastOplogs(logs []*Oplog, msg OpType, pendingMsg OpType) error {
	lenOplog := len(logs)

	// extras
	origExtras := make([]interface{}, lenOplog)
	for i, log := range logs {
		origExtras[i] = log.Extra
	}
	defer func() {
		for i, log := range logs {
			log.Extra = origExtras[i]
		}
	}()

	for _, log := range logs {
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
	meLogs := make([]*Oplog, 0, lenOplog)
	masterLogs := make([]*Oplog, 0, lenOplog)
	allLogs := make([]*Oplog, 0, lenOplog)

	for _, log := range logs {
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
		pm.SendDataToPeers(pendingMsg, meLogs, mePeerList)
	}

	if len(masterLogs) != 0 && len(masterPeerList) != 0 {
		pm.SendDataToPeers(pendingMsg, masterLogs, masterPeerList)
	}

	if len(allLogs) != 0 && len(allPeerList) != 0 {
		pm.SendDataToPeers(msg, allLogs, allPeerList)
	}

	return nil
}

func (pm *BaseProtocolManager) InternalSign(log *Oplog) (bool, error) {
	if log.MasterLogID != nil {
		return false, nil
	}

	ptt := pm.Ptt()
	myEntity := ptt.MyEntity()
	myID := myEntity.GetID()

	// check
	if !reflect.DeepEqual(myID, log.DoerID) && !pm.IsMaster(myID) {
		return false, nil
	}

	masterSigns := log.MasterSigns
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

	key := myEntity.SignKey()
	nodeSignID := myEntity.GetNodeSignID()

	internalSigns := log.InternalSigns
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
	err := log.InternalSign(nodeSignID, key)
	if err != nil {
		return false, err
	}

	_, weight, isValid := myEntity.IsValidInternalOplog(log.InternalSigns)
	if !isValid {
		return true, nil
	}

	// master-sign
	err = log.MasterSign(myID, key)
	if err != nil {
		return false, err
	}

	masterLogID, weight, isValid := pm.isValidOplog(log.MasterSigns)
	if !isValid {
		return true, nil
	}

	// master-log-id
	err = log.SetMasterLogID(masterLogID, weight)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pm *BaseProtocolManager) GetPendingOplogs(setDB func(*Oplog)) ([]*Oplog, []*Oplog, error) {
	log := &Oplog{}
	setDB(log)

	expireTime, err := types.GetTimestamp()
	if err != nil {
		return nil, nil, err
	}
	expireTime.Ts -= ExpireOplogSeconds

	pendingLogs, err := pm.GetOplogList(log, nil, 0, pttdb.ListOrderNext, types.StatusPending, false)
	if err != nil {
		return nil, nil, err
	}

	internalPendingLogs, err := pm.GetOplogList(log, nil, 0, pttdb.ListOrderNext, types.StatusInternalPending, false)
	if err != nil {
		return nil, nil, err
	}

	lenLogs := len(pendingLogs) + len(internalPendingLogs)
	logs := make([]*Oplog, 0, lenLogs)
	failedLogs := make([]*Oplog, 0, lenLogs)

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

func (pm *BaseProtocolManager) IntegrateOplog(log *Oplog, isLocked bool) (bool, error) {
	if !isLocked {
		err := log.Lock()
		if err != nil {
			return false, err
		}
		defer log.Unlock()
	}

	isToSign, err := log.IntegrateExisting(true)
	if err != nil {
		return false, err
	}
	if !isToSign {
		return false, nil
	}

	err = pm.validateIntegrateSign(log, true)
	if err != nil {
		return false, err
	}

	err = log.Save(true)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (pm *BaseProtocolManager) validateIntegrateSign(log *Oplog, isLocked bool) error {
	var err error
	if !isLocked {
		err = log.Lock()
		if err != nil {
			return err
		}
		defer log.Unlock()
	}

	ptt := pm.Ptt()
	myEntity := ptt.MyEntity()
	myID := myEntity.GetID()
	key := ptt.SignKey()

	_, weight, isValid := myEntity.IsValidInternalOplog(log.InternalSigns)
	if isValid {
		err = log.MasterSign(myID, key)
		if err != nil {
			return err
		}
	}

	masterLogID, weight, isValid := pm.isValidOplog(log.MasterSigns)
	if isValid {
		err = log.SetMasterLogID(masterLogID, weight)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *BaseProtocolManager) RemoveNonSyncOplog(setDB func(log *Oplog), logID *types.PttID, isRetainValid bool, isLocked bool) (*Oplog, error) {
	log := &Oplog{}
	setDB(log)
	log.ID = logID

	if !isLocked {
		err := log.Lock()
		if err != nil {
			return nil, err
		}
		defer log.Unlock()
	}

	err := log.Get(logID, true)
	if err != nil {
		return nil, err
	}

	status := log.ToStatus()
	if log.IsSync && status == types.StatusAlive {
		return nil, nil
	}

	if isRetainValid && status == types.StatusAlive {
		log.IsSync = true
		err = log.SaveWithIsSync(true)
		return log, nil
	}

	err = log.Delete(true)

	return nil, err
}

func (pm *BaseProtocolManager) SetOplogIsSync(log *Oplog) (bool, error) {
	log.IsSync = true
	isNewSign, err := pm.InternalSign(log)
	if err != nil {
		return false, err
	}

	return isNewSign, nil
}
