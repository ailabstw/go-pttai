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
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

/**********
 * AddMasterOplog
 **********/

func (pm *BaseProtocolManager) HandleAddMasterOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddOplog(dataBytes, pm.HandleMasterOplogs, peer)
}

func (pm *BaseProtocolManager) HandleAddMasterOplogs(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddOplogs(dataBytes, pm.HandleMasterOplogs, peer)
}

func (pm *BaseProtocolManager) HandleAddPendingMasterOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddPendingOplog(dataBytes, pm.HandlePendingMasterOplogs, peer)
}

func (pm *BaseProtocolManager) HandleAddPendingMasterOplogs(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleAddPendingOplogs(dataBytes, pm.HandlePendingMasterOplogs, peer)
}

/**********
 * SyncMasterOplog
 **********/

func (pm *BaseProtocolManager) HandleSyncMasterOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncOplog(
		dataBytes,
		peer,

		pm.MasterMerkle(),

		ForceSyncMasterOplogMsg,
		ForceSyncMasterOplogAckMsg,
		InvalidSyncMasterOplogMsg,
		SyncMasterOplogAckMsg,
	)
}

func (pm *BaseProtocolManager) HandleForceSyncMasterOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleForceSyncOplog(
		dataBytes,
		peer,

		pm.MasterMerkle(),
		ForceSyncMasterOplogAckMsg,
	)
}

func (pm *BaseProtocolManager) HandleForceSyncMasterOplogAck(dataBytes []byte, peer *PttPeer) error {

	info := NewProcessPersonInfo()

	return pm.HandleForceSyncOplogAck(
		dataBytes,
		peer,

		pm.MasterMerkle(),
		info,

		pm.SetMasterDB,
		pm.HandleFailedValidMasterOplog,
		pm.SetNewestMasterOplog,
		pm.postprocessFailedValidMasterOplogs,

		SyncMasterOplogNewOplogsMsg,
	)
}

func (pm *BaseProtocolManager) HandleSyncMasterOplogInvalidAck(dataBytes []byte, peer *PttPeer) error {

	return pm.HandleSyncOplogInvalidAck(
		dataBytes,
		peer,

		pm.MasterMerkle(),
		ForceSyncMasterOplogMsg,
	)
}

func (pm *BaseProtocolManager) HandleSyncMasterOplogAck(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncOplogAck(
		dataBytes,
		peer,

		pm.MasterMerkle(),

		pm.SetMasterDB,
		pm.SetNewestMasterOplog,
		pm.postsyncMasterOplogs,

		SyncMasterOplogNewOplogsMsg,
	)
}

func (pm *BaseProtocolManager) HandleSyncNewMasterOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncOplogNewOplogs(
		dataBytes,
		peer,

		pm.SetMasterDB,
		pm.HandleMasterOplogs,
		pm.SetNewestMasterOplog,

		SyncMasterOplogNewOplogsAckMsg,
	)
}

func (pm *BaseProtocolManager) HandleSyncNewMasterOplogAck(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncOplogNewOplogsAck(
		dataBytes,
		peer,

		pm.SetMasterDB,
		pm.HandleMasterOplogs,
		pm.postsyncMasterOplogs,
	)
}

/**********
 * SyncPendingMasterOplog
 **********/

func (pm *BaseProtocolManager) HandleSyncPendingMasterOplog(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncPendingOplog(
		dataBytes,
		peer,

		pm.HandlePendingMasterOplogs,
		pm.SetMasterDB,
		pm.HandleFailedMasterOplog,

		SyncPendingMasterOplogAckMsg,
	)
}

func (pm *BaseProtocolManager) HandleSyncPendingMasterOplogAck(dataBytes []byte, peer *PttPeer) error {
	return pm.HandleSyncPendingOplogAck(
		dataBytes,
		peer,

		pm.HandlePendingMasterOplogs,
	)
}

/**********
 * HandleOplogs
 **********/

func (pm *BaseProtocolManager) HandleMasterOplogs(oplogs []*BaseOplog, peer *PttPeer, isUpdateSyncTime bool) error {

	info := NewProcessPersonInfo()

	return HandleOplogs(
		oplogs,
		peer,

		isUpdateSyncTime,

		pm,
		info,

		pm.masterMerkle,

		pm.SetMasterDB,
		pm.processMasterLog,
		pm.postprocessMasterOplogs,
	)

}

func (pm *BaseProtocolManager) HandlePendingMasterOplogs(oplogs []*BaseOplog, peer *PttPeer) error {

	info := NewProcessPersonInfo()

	oplogs, err := preprocessOplogs(oplogs, pm.SetMasterDB, false, pm, nil, peer)
	if err != nil {
		return err
	}

	return pm.handlePendingMasterOplogs(
		oplogs,
		peer,

		info,

		pm.processPendingMasterLog,
		pm.processMasterLog,
	)

}

func (pm *BaseProtocolManager) handlePendingMasterOplogs(
	oplogs []*BaseOplog,
	peer *PttPeer,

	info ProcessInfo,

	processPendingLog func(oplog *BaseOplog, i ProcessInfo) (types.Bool, []*BaseOplog, error),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
) error {

	var err error
	var origLogs []*BaseOplog

	isToBroadcast := false
	toBroadcastLogs := make([]*BaseOplog, 0, len(oplogs))
	for _, oplog := range oplogs {
		isToBroadcast, origLogs, err = pm.handlePendingMasterOplog(oplog, info, processPendingLog, processLog)
		if err == ErrSkipOplog {
			continue
		}
		if err != nil {
			break
		}

		if len(origLogs) != 0 {
			toBroadcastLogs = append(toBroadcastLogs, origLogs...)
		}

		// new-sign
		if isToBroadcast {
			toBroadcastLogs = append(toBroadcastLogs, oplog)
		}
	}

	pm.broadcastMasterOplogsCore(toBroadcastLogs)

	return err
}

/*
HandlePendingOplog Handles single pending oplog.
    1. lock oplog.
    2. integrate oplog.
    3. process pending board oplog
    4. sync and internal-sign.
    5. save-with-is-sync
*/
func (pm *BaseProtocolManager) handlePendingMasterOplog(
	oplog *BaseOplog,

	info ProcessInfo,

	processPendingLog func(oplog *BaseOplog, i ProcessInfo) (types.Bool, []*BaseOplog, error),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
) (bool, []*BaseOplog, error) {

	err := oplog.Lock()
	if err != nil {
		return false, nil, err
	}
	defer oplog.Unlock()

	merkle := pm.MasterMerkle()

	// integrate
	// after integrate-me-oplog: oplog saved if orig exists and not new-signed.
	isNewSign, err := pm.IntegrateOplog(oplog, true, merkle)
	if err != nil {
		return false, nil, err
	}

	if oplog.IsSync {
		if isNewSign {
			err = oplog.Save(true, merkle)
			if err != nil {
				return false, nil, err
			}
		}
		return isNewSign, nil, nil
	}

	var fromID *types.PttID
	var toID *types.PttID

	if oplog.Op == MasterOpTypeTransferMaster {
		fromID = oplog.ObjID
		opData := &PersonOpTransferPerson{}
		err = oplog.GetData(opData)
		if err != nil {
			return false, nil, err
		}
		toID = opData.ToID

		if oplog.MasterLogID != nil {
			err = pm.checkTransferMasterSign(oplog, fromID, toID)
			if err != nil {
				return false, nil, err
			}
		}
	}

	// master-oplog-id after integrated
	if oplog.MasterLogID != nil {
		// process log
		origLogs, err := processLog(oplog, info)
		if err == ErrNewerOplog {
			oplog.IsSync = true
			err = nil
		}

		if err != nil {
			return false, nil, err
		}

		return true, origLogs, nil
	}

	// process pending log
	isToSign, origLogs, err := processPendingLog(oplog, info)
	if err == ErrNewerOplog {
		err = ErrSkipOplog
	}
	if err != nil {
		return false, nil, err
	}

	// is-sync: sign
	if isToSign {
		if oplog.Op == MasterOpTypeTransferMaster {
			err = pm.signMasterOplog(oplog, fromID, toID)
			if err != nil {
				return false, nil, err
			}
		} else {
			_, err = pm.InternalSign(oplog)
			if err != nil {
				return false, nil, err
			}
		}

		isNewSign = true
	}

	// save oplog
	err = oplog.SaveWithIsSync(true)
	if err == pttdb.ErrInvalidUpdateTS {
		err = nil
	}
	if err != nil {
		return false, nil, err
	}

	return isNewSign, origLogs, nil
}
