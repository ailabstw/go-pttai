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
HandleOplogs handles a list of MeOplog.
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
	oplogs []*BaseOplog, peer *PttPeer,
	isUpdateSyncTime bool,

	info ProcessInfo,
	merkle *Merkle,

	setDB func(oplog *BaseOplog),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
	postprocessLogs func(i ProcessInfo, origLogs []*BaseOplog, p *PttPeer, isPending bool) error,
) error {

	var err error

	oplogs, err = preprocessOplogs(oplogs, setDB)
	if err != nil {
		return err
	}

	// handle oplogs
	newestUpdateTS, err := handleOplogs(oplogs, peer, info, processLog, postprocessLogs)

	// update-sync-time
	var err2 error
	if isUpdateSyncTime && merkle != nil {
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
	oplogs []*BaseOplog, peer *PttPeer,

	info ProcessInfo,

	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
	postprocessLogs func(i ProcessInfo, toBroadcastLogs []*BaseOplog, p *PttPeer, isPending bool) error,
) (types.Timestamp, error) {

	// handleOplogs
	var err error
	var origLogs []*BaseOplog
	var newestUpdateTS types.Timestamp

	isToBroadcast := false
	toBroadcastLogs := make([]*BaseOplog, 0, len(oplogs))
	for _, oplog := range oplogs {
		isToBroadcast, origLogs, err = handleOplog(oplog, info, processLog)
		log.Debug("handleOplogs: after handleOplog", "isToBroadcast", isToBroadcast, "origLogs", origLogs, "e", err)

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
	oplog *BaseOplog, info ProcessInfo,

	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
) (bool, []*BaseOplog, error) {

	err := oplog.Lock()
	if err != nil {
		return false, nil, err
	}
	defer oplog.Unlock()

	// select
	err = oplog.SelectExisting(true)
	if err != nil {
		return false, nil, err
	}
	if oplog.IsSync {
		return false, nil, nil
	}

	// process log
	origLogs, err := processLog(oplog, info)
	log.Debug("handleOplog: after processLog", "oplog", oplog, "e", err, "IsSync", oplog.IsSync)
	isSync := bool(oplog.IsSync)
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
	oplogs []*BaseOplog, peer *PttPeer,

	pm ProtocolManager,
	info ProcessInfo,

	setDB func(oplog *BaseOplog),
	processPendingLog func(oplog *BaseOplog, i ProcessInfo) ([]*BaseOplog, error),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
	postprocessLogs func(i ProcessInfo, toBroadcastLogs []*BaseOplog, p *PttPeer, isPending bool) error,
) error {

	var err error

	oplogs, err = preprocessOplogs(oplogs, setDB)
	if err != nil {
		return err
	}

	// process
	err = handlePendingOplogs(oplogs, peer, pm, info, processPendingLog, processLog, postprocessLogs)
	if err != nil {
		return err
	}

	return nil
}

func handlePendingOplogs(
	oplogs []*BaseOplog, peer *PttPeer,

	pm ProtocolManager,
	info ProcessInfo,

	processPendingLog func(oplog *BaseOplog, i ProcessInfo) ([]*BaseOplog, error),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
	postprocessLogs func(i ProcessInfo, toBroadcastLogs []*BaseOplog, p *PttPeer, isPending bool) error,
) error {

	var err error
	var origLogs []*BaseOplog

	isToBroadcast := false
	toBroadcastLogs := make([]*BaseOplog, 0, len(oplogs))
	for _, oplog := range oplogs {
		isToBroadcast, origLogs, err = handlePendingOplog(oplog, pm, info, processPendingLog, processLog)
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
	oplog *BaseOplog, pm ProtocolManager,
	info ProcessInfo,

	processPendingLog func(oplog *BaseOplog, i ProcessInfo) ([]*BaseOplog, error),
	processLog func(oplog *BaseOplog, info ProcessInfo) ([]*BaseOplog, error),
) (bool, []*BaseOplog, error) {

	err := oplog.Lock()
	if err != nil {
		return false, nil, err
	}
	defer oplog.Unlock()

	// integrate
	// after integrate-me-oplog: oplog saved if orig exists and not new-signed.
	isNewSign, err := pm.IntegrateOplog(oplog, true)
	if err != nil {
		return false, nil, err
	}

	if oplog.IsSync {
		if isNewSign {
			err = oplog.Save(true)
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
	origLogs, err := processPendingLog(oplog, info)
	if err == ErrNewerOplog {
		err = ErrSkipOplog
	}
	if err != nil {
		return false, nil, err
	}

	// is-sync: sign
	if oplog.IsSync {
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
	oplogs []*BaseOplog, setDB func(oplog *BaseOplog),
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

func preprocessOplogs(oplogs []*BaseOplog, setDB func(oplog *BaseOplog)) ([]*BaseOplog, error) {
	var err error

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
		if err != nil {
			badIdx = i
			break
		}
	}

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
