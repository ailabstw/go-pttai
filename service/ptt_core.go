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

func (p *Ptt) GetVersion() (string, error) {
	return p.config.Version, nil
}

func (p *Ptt) GetGitCommit() (string, error) {
	return p.config.GitCommit, nil
}

func (p *Ptt) CountPeers() (*BackendCountPeers, error) {
	p.peerLock.RLock()
	defer p.peerLock.RUnlock()

	return &BackendCountPeers{
		MyPeers:        len(p.myPeers),
		ImportantPeers: len(p.importantPeers),
		MemberPeers:    len(p.memberPeers),
		RandomPeers:    len(p.randomPeers),
	}, nil
}

func (p *Ptt) BEGetPeers() ([]*BackendPeer, error) {
	p.peerLock.RLock()
	defer p.peerLock.RUnlock()

	peerList := make([]*BackendPeer, 0, len(p.myPeers)+len(p.importantPeers)+len(p.memberPeers)+len(p.randomPeers))

	var backendPeer *BackendPeer
	for _, peer := range p.myPeers {
		backendPeer = PeerToBackendPeer(peer)
		peerList = append(peerList, backendPeer)
	}

	for _, peer := range p.importantPeers {
		backendPeer = PeerToBackendPeer(peer)
		peerList = append(peerList, backendPeer)
	}

	for _, peer := range p.memberPeers {
		backendPeer = PeerToBackendPeer(peer)
		peerList = append(peerList, backendPeer)
	}

	for _, peer := range p.randomPeers {
		backendPeer = PeerToBackendPeer(peer)
		peerList = append(peerList, backendPeer)
	}

	return peerList, nil
}

func (p *Ptt) Shutdown() (bool, error) {
	p.NotifyNodeStop.PassChan(struct{}{})
	return false, nil
}

func (p *Ptt) Restart() (bool, error) {
	p.NotifyNodeRestart.PassChan(struct{}{})
	return false, nil
}
