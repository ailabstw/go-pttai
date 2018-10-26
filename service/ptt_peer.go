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
	"sync"
	"time"

	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/p2p"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

type PttPeer struct {
	*p2p.Peer

	LockPeerType sync.RWMutex
	PeerType     PeerType

	rw p2p.MsgReadWriter

	version uint

	term chan struct{} // Termination channel to stop the broadcaster

	ptt *BasePtt

	UserID *types.PttID

	lockID      sync.Mutex
	IDEntityID  *types.PttID
	IDChallenge *types.Salt
	IDChan      chan struct{}
}

func NewPttPeer(version uint, p *p2p.Peer, rw p2p.MsgReadWriter, ptt *BasePtt) (*PttPeer, error) {
	return &PttPeer{
		Peer:    p,
		rw:      rw,
		version: version,
		ptt:     ptt,

		term:   make(chan struct{}),
		IDChan: make(chan struct{}, 1),
	}, nil
}

func (p *PttPeer) GetID() *discover.NodeID {
	id := p.Peer.ID()
	return &id
}

func (p *PttPeer) Broadcast() error {
	return nil
}

func (p *PttPeer) Info() interface{} {
	return &PttPeerInfo{
		NodeID:   p.GetID(),
		UserID:   p.UserID,
		PeerType: p.PeerType,
	}
}

func (p *PttPeer) Handshake(networkID uint32) error {
	errc := make(chan error, 2)

	go func() {
		errc <- p2p.Send(p.rw, uint64(CodeTypeStatus), &PttStatus{
			Version:   uint32(p.version),
			NetworkID: networkID,
		})
	}()

	go func() {
		errc <- p.ReadStatus(networkID)
	}()

	timeout := time.NewTimer(HandshakeTimeout)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errc:
			if err != nil {
				return err
			}
		case <-timeout.C:
			return p2p.DiscReadTimeout
		}
	}

	return nil
}

func (p *PttPeer) ReadStatus(networkID uint32) error {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}

	if msg.Code != uint64(CodeTypeStatus) {
		return ErrInvalidMsgCode
	}

	if msg.Size > ProtocolMaxMsgSize {
		return ErrMsgTooLarge
	}

	status := &PttStatus{}
	err = msg.Decode(&status)
	if err != nil {
		return err
	}

	if status.NetworkID != networkID {
		return ErrInvalidData
	}

	if uint(status.Version) != p.version {
		return ErrInvalidData
	}

	return nil
}

func (p *PttPeer) GetPeer() *p2p.Peer {
	return p.Peer
}

func (p *PttPeer) Version() uint {
	return p.version
}

func (p *PttPeer) RW() p2p.MsgReadWriter {
	return p.rw
}

func (p *PttPeer) SendData(data *PttData) error {
	//log.Debug("SendData", "p", p, "data", data)
	return p2p.Send(p.rw, uint64(data.Code), data)
}

/**********
 * Identify UserID
 **********/

/*
InitID initializes info for identifying user-id
*/
func (p *PttPeer) InitID(entityID *types.PttID, salt *types.Salt, quitSync chan struct{}) error {
	p.lockID.Lock()
	defer p.lockID.Unlock()

	if p.UserID != nil {
		return types.ErrAlreadyExists
	}

	if p.IDEntityID != nil {
		return types.ErrAlreadyExists
	}
	p.IDEntityID = entityID
	p.IDChallenge = salt

	// expire-time
	go func(p *PttPeer, entityID *types.PttID) {
		timer := time.NewTimer(IdentifyPeerTimeout)
		defer timer.Stop()

		select {
		case <-timer.C:
			p.FinishID(entityID)
		case <-quitSync:
			break
		}
	}(p, entityID)

	return nil
}

/*
FinishID finishes info for identifying user-id (remove info)
*/
func (p *PttPeer) FinishID(entityID *types.PttID) {
	p.lockID.Lock()
	defer p.lockID.Unlock()

	if !reflect.DeepEqual(entityID, p.IDEntityID) {
		return
	}

	p.IDEntityID = nil
}
