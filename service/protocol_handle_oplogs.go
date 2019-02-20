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
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/pttdb"
)

/*
HandleOplogs handles a list of Oplog.
    1. verify all the oplogs, return if any of the log is invalid.
    2. preset oplog (isSync as false and setDB)
    3. check pre-log-id
    4. integrate existing oplog. skip if already synced.
    5. process me-oplog.
    6. save me-oplog.
    7. broadcast me-oplogs.
    8. save sync-time.
*/
func HandleOplogs(
	oplogs []*BaseOplog,
	peer *PttPeer,

	isUpdateSyncTime bool,

	pm ProtocolManager,
	info ProcessInfo,

	merkle *Merkle,

	setDB func(oplog *BaseOplog),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
	postprocessLogs func(i ProcessInfo, origLogs []*BaseOplog, p *PttPeer, isPending bool) error,
) error {

	var err error

	if len(oplogs) == 0 {
		return nil
	}

	oplogs, err = preprocessOplogs(oplogs, setDB, isUpdateSyncTime, pm, merkle, peer)
	log.Debug("HandleOplogs: after preprocessOplogs", "e", err, "oplogs", oplogs, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
	if err != nil {
		return err
	}
	if len(oplogs) == 0 {
		return nil
	}

	// handle oplogs
	newestUpdateTS, err := handleOplogs(
		oplogs,
		peer,

		pm,
		info,

		merkle,

		processLog,
		postprocessLogs,
	)

	// update-sync-time
	var err2 error
	if isUpdateSyncTime && merkle != nil {
		log.Debug("HandleOplogs: to save sync time", "ts", newestUpdateTS, "entity", pm.Entity().GetID(), "service", pm.Entity().Service().Name())
		err2 = merkle.SaveSyncTime(newestUpdateTS)
		if err2 == pttdb.ErrInvalidUpdateTS {
			err2 = nil
		}
	}

	if err != nil {
		log.Error("HandleOplogs: unable to process oplog", "e", err)
		return err
	}

	if err2 != nil {
		log.Error("HandleOplogs: unable to save sync-time", "e", err2)
		return err2
	}

	return nil
}

func handleOplogs(
	oplogs []*BaseOplog,
	peer *PttPeer,

	pm ProtocolManager,
	info ProcessInfo,

	merkle *Merkle,

	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
	postprocessLogs func(i ProcessInfo, toBroadcastLogs []*BaseOplog, p *PttPeer, isPending bool) error,
) (types.Timestamp, error) {

	// handleOplogs
	var err error
	var origLogs []*BaseOplog
	var newestUpdateTS types.Timestamp

	isToBroadcast := types.Bool(false)
	toBroadcastLogs := make([]*BaseOplog, 0, len(oplogs))
	for _, oplog := range oplogs {
		log.Debug("handleOplogs (in-for-loop): to handleOplog", "updateTS", oplog.UpdateTS, "entity", pm.Entity().GetID())
		isToBroadcast, origLogs, err = handleOplog(
			oplog,
			info,

			merkle,

			processLog,
		)
		log.Debug("handleOplogs: after handleOplog", "isToBroadcast", isToBroadcast, "origLogs", origLogs, "updateTS", oplog.UpdateTS, "e", err, "entity", pm.Entity().GetID())
		if err == ErrSkipOplog {
			continue
		}
		if err != nil {
			break
		}

		if len(origLogs) != 0 {
			toBroadcastLogs = append(toBroadcastLogs, origLogs...)
		}

		if isToBroadcast {
			toBroadcastLogs = append(toBroadcastLogs, oplog)
		}

		if newestUpdateTS.IsLess(oplog.UpdateTS) {
			newestUpdateTS = oplog.UpdateTS
		}
	}

	postprocessLogs(info, toBroadcastLogs, peer, false)

	return newestUpdateTS, err
}

func handleOplog(
	oplog *BaseOplog,
	info ProcessInfo,

	merkle *Merkle,

	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
) (types.Bool, []*BaseOplog, error) {

	err := oplog.Lock()
	if err != nil {
		return false, nil, err
	}
	defer oplog.Unlock()

	// select
	isToBroadcast, err := oplog.SelectExisting(true, merkle)
	log.Debug("handleOplog: after SelectExisting", "oplog", oplog, "e", err, "IsSync", oplog.IsSync)
	if err != nil {
		return false, nil, err
	}
	if oplog.IsSync {
		return isToBroadcast, nil, nil
	}

	// process log
	origLogs, err := processLog(oplog, info)
	log.Debug("handleOplog: after processLog", "oplog", oplog, "e", err, "IsSync", oplog.IsSync)
	isSync := oplog.IsSync
	if err == ErrNewerOplog {
		oplog.IsSync = true
		err = nil
	}

	if err != nil {
		return false, nil, err
	}

	// save oplog
	err = oplog.SaveWithIsSync(true)
	if err != nil && err != pttdb.ErrInvalidUpdateTS {
		return false, nil, err
	}

	return isSync, origLogs, nil
}

/**********
 * Handle Pending Oplogs
 **********/

func HandlePendingOplogs(
	oplogs []*BaseOplog,
	peer *PttPeer,

	pm ProtocolManager,
	info ProcessInfo,

	merkle *Merkle,

	setDB func(oplog *BaseOplog),
	processPendingLog func(oplog *BaseOplog, i ProcessInfo) (types.Bool, []*BaseOplog, error),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
	postprocessLogs func(i ProcessInfo, toBroadcastLogs []*BaseOplog, p *PttPeer, isPending bool) error,
) error {

	var err error

	if len(oplogs) == 0 {
		return nil
	}

	log.Debug("HandlePendingOplogs: to preprocessOplogs", "e", err, "oplogs", oplogs)
	oplogs, err = preprocessOplogs(oplogs, setDB, false, pm, nil, peer)
	log.Debug("HandlePendingOplogs: after preprocessOplogs", "e", err, "oplogs", oplogs)
	if err != nil {
		return err
	}

	if len(oplogs) == 0 {
		return nil
	}

	// process
	err = handlePendingOplogs(
		oplogs,
		peer,

		pm,
		info,

		merkle,

		processPendingLog,
		processLog,
		postprocessLogs,
	)
	log.Debug("HandlePendingOplogs: after handlePendingOplogs", "e", err)
	if err != nil {
		return err
	}

	return nil
}

/*

We require separated isToSign from processPendingLog, because in the "delete" situation, we still need to sign the oplog, but the oplog is not synced yet.
*/
func handlePendingOplogs(
	oplogs []*BaseOplog,
	peer *PttPeer,

	pm ProtocolManager,
	info ProcessInfo,

	merkle *Merkle,

	processPendingLog func(oplog *BaseOplog, i ProcessInfo) (types.Bool, []*BaseOplog, error),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
	postprocessLogs func(i ProcessInfo, toBroadcastLogs []*BaseOplog, p *PttPeer, isPending bool) error,
) error {

	var err error
	var origLogs []*BaseOplog

	isToBroadcast := false
	toBroadcastLogs := make([]*BaseOplog, 0, len(oplogs))
	for _, oplog := range oplogs {
		isToBroadcast, origLogs, err = handlePendingOplog(
			oplog,
			pm,
			info,

			merkle,

			processPendingLog,
			processLog,
		)
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
		log.Debug("handlePendingOplogs (in-for-loop)", "isToBroadcast", isToBroadcast, "entity", pm.Entity().GetID())
		if isToBroadcast {
			toBroadcastLogs = append(toBroadcastLogs, oplog)
		}
	}

	postprocessLogs(info, toBroadcastLogs, peer, true)

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
func handlePendingOplog(
	oplog *BaseOplog,
	pm ProtocolManager,
	info ProcessInfo,

	merkle *Merkle,

	processPendingLog func(oplog *BaseOplog, i ProcessInfo) (types.Bool, []*BaseOplog, error),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
) (bool, []*BaseOplog, error) {

	err := oplog.Lock()
	if err != nil {
		return false, nil, err
	}
	defer oplog.Unlock()

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
		_, err = pm.InternalSign(oplog)
		if err != nil {
			return false, nil, err
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

/**********
 * Handle Failed Oplogs
 **********/

func HandleFailedOplogs(
	oplogs []*BaseOplog,

	setDB func(oplog *BaseOplog),
	handleFailedOplog func(oplog *BaseOplog) error,
) error {

	var err error
	for _, oplog := range oplogs {
		setDB(oplog)

		err = handleFailedOplog(oplog)
		if err != nil {
			continue
		}

		oplog.Delete(false)
	}

	return nil
}

func HandleFailedValidOplogs(
	oplogs []*BaseOplog,
	peer *PttPeer,

	info ProcessInfo,

	setDB func(oplog *BaseOplog),
	handleFailedValidOplog func(oplog *BaseOplog, info ProcessInfo) error,
	postprocessLogs func(info ProcessInfo, peer *PttPeer) error,
) error {

	var err error
	for _, oplog := range oplogs {
		setDB(oplog)

		err = handleFailedValidOplog(oplog, info)
		if err != nil {
			continue
		}

		oplog.Delete(false)
	}

	postprocessLogs(info, peer)

	return nil
}

func preprocessOplogs(
	oplogs []*BaseOplog,
	setDB func(oplog *BaseOplog),
	isUpdateSyncTime bool,

	pm ProtocolManager,
	merkle *Merkle,
	peer *PttPeer,
) ([]*BaseOplog, error) {
	var err error

	// expire-ts
	now, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}
	expireTS := types.ZeroTimestamp
	if merkle != nil {
		ts, err := merkle.GetSyncTime()
		if err != nil {
			return nil, err
		}
		expireTS = ts
	}
	if expireTS.Ts >= OffsetMerkleSyncTime {
		expireTS.Ts -= OffsetMerkleSyncTime
	} else {
		expireTS = types.ZeroTimestamp
	}

	// expire-ts: start-idx
	startIdx := len(oplogs)
	for i, oplog := range oplogs {
		if expireTS.IsLess(oplog.UpdateTS) {
			startIdx = i
			break
		}
	}
	if startIdx != 0 {
		expiredLog := oplogs[0]
		log.Warn("preprocessOplogs: received expired oplogs", "e", pm.Entity().GetID(), "expiredLog", expiredLog.ID, "expiredTS", expiredLog.UpdateTS, "expireTS", expireTS, "peer", peer)
		oplogs = oplogs[startIdx:]
	}

	log.Debug("preprocessOplogs: after startIdx", "startIdx", startIdx, "oplogs", oplogs, "entity", pm.Entity().GetID())

	// future-ts: end-idx
	lenLogs := len(oplogs)
	endIdx := 0
	for i := lenLogs - 1; i >= 0; i-- {
		if oplogs[i].UpdateTS.IsLess(now) {
			endIdx = i + 1
			break
		}
	}
	if endIdx != lenLogs {
		futureLog := oplogs[lenLogs-1]
		log.Warn("preprocessOplogs: received future oplogs", "futureLog", futureLog.ID, "futureTS", futureLog.UpdateTS, "now", now, "peer", peer)
	}
	oplogs = oplogs[:endIdx]

	log.Debug("preprocessOplogs: after endIdx", "endIdx", endIdx, "oplogs", len(oplogs), "entity", pm.Entity().GetID())

	if len(oplogs) == 0 {
		return oplogs, nil
	}

	// init
	for _, oplog := range oplogs {
		oplog.IsSync = false
		setDB(oplog)
	}

	// verify
	// return err if any of the oplog is invalid
	for _, oplog := range oplogs {
		if oplog == nil {
			return nil, ErrInvalidOplog
		}

		err = oplog.Verify()
		if err != nil {
			log.Debug("preprocessOplogs: unable to verify oplog", "op", oplog.Op, "e", err)
			return nil, err
		}
	}

	// check pre-log-id
	// XXX prelog as shared tmp-variable.
	prelog := &BaseOplog{}
	setDB(prelog)
	existIDs := make(map[types.PttID]*BaseOplog)
	badIdx := len(oplogs)
	for i, oplog := range oplogs {
		err = checkPreOplog(oplog, prelog, existIDs)
		log.Debug("preprocessOplogs: (in-for-loop) after checkPreOplog", "i", i, "e", err, "preLogID", oplog.PreLogID)
		if err != nil {
			badIdx = i
			break
		}
	}

	log.Debug("preprocessOplogs: after for-loop", "badIdx", badIdx)

	return oplogs[:badIdx], nil
}

func checkPreOplog(oplog *BaseOplog, prelog *BaseOplog, existIDs map[types.PttID]*BaseOplog) error {

	if oplog.PreLogID == nil {
		existIDs[*oplog.ID] = oplog
		return nil
	}

	log, ok := existIDs[*oplog.PreLogID]
	if ok && log.MasterLogID != nil {
		existIDs[*oplog.ID] = oplog
		return nil
	}

	err := prelog.Get(oplog.PreLogID, false)
	if err != nil {
		return ErrInvalidOplog
	}

	if prelog.MasterLogID == nil {
		return ErrInvalidOplog
	}

	existIDs[*oplog.ID] = oplog
	return nil
}
