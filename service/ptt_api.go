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
	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/pttdb"
)

type PrivateAPI struct {
	p *BasePtt
}

func NewPrivateAPI(p *BasePtt) *PrivateAPI {
	return &PrivateAPI{p}
}

func (api *PrivateAPI) GetVersion() (string, error) {
	return api.p.GetVersion()
}

func (api *PrivateAPI) GetGitCommit() (string, error) {
	return api.p.GetGitCommit()
}

func (api *PrivateAPI) Shutdown() (bool, error) {
	return api.p.Shutdown()
}

func (api *PrivateAPI) Restart() (bool, error) {
	return api.p.Restart()
}

/**********
 * Peer
 **********/

func (api *PrivateAPI) CountPeers() (*BackendCountPeers, error) {
	return api.p.CountPeers()
}

func (api *PrivateAPI) GetPeers() ([]*BackendPeer, error) {
	return api.p.BEGetPeers()
}

/**********
 * Entities
 **********/

func (api *PrivateAPI) CountEntities() (int, error) {
	return api.p.CountEntities()
}

/**********
 * Join
 **********/

func (api *PrivateAPI) GetJoins() (map[common.Address]*types.PttID, error) {
	return api.p.GetJoins(), nil

}

func (api *PrivateAPI) GetConfirmJoins() ([]*BackendConfirmJoin, error) {
	return api.p.GetConfirmJoins()
}

/**********
 * Op
 **********/

func (api *PrivateAPI) GetOps() (map[common.Address]*types.PttID, error) {
	return api.p.GetOps(), nil
}

/**********
 * PttOplog
 **********/

func (api *PrivateAPI) GetPttOplogList(logID string, limit int, listOrder pttdb.ListOrder) ([]*PttOplog, error) {
	return api.p.BEGetPttOplogList([]byte(logID), limit, listOrder)
}

func (api *PrivateAPI) MarkPttOplogSeen() (types.Timestamp, error) {
	return api.p.MarkPttOplogSeen()
}

func (api *PrivateAPI) GetPttOplogSeen() (types.Timestamp, error) {
	return api.p.GetPttOplogSeen()
}

/**********
 * Locale
 **********/

func (api *PrivateAPI) SetLocale(locale Locale) (Locale, error) {
	err := SetLocale(locale)
	return CurrentLocale, err
}

func (api *PrivateAPI) GetLocale() (Locale, error) {
	return CurrentLocale, nil
}

/**********
 * P2P
 **********/

func (api *PrivateAPI) GetLastAnnounceP2PTS() (types.Timestamp, error) {
	return api.p.GetLastAnnounceP2PTS()
}
