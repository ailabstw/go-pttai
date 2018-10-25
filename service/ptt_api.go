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

type PrivateAPI struct {
	p *BasePtt
}

func NewPrivateAPI(p *BasePtt) *PrivateAPI {
	return &PrivateAPI{p}
}

func (api *PrivateAPI) CountPeers() (*BackendCountPeers, error) {
	return api.p.CountPeers()
}

func (api *PrivateAPI) GetPeers() ([]*BackendPeer, error) {
	return api.p.BEGetPeers()
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
