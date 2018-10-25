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

	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

/*
HandleMessageWrapper
*/
func (p *BasePtt) HandleMessageWrapper(peer *PttPeer) error {
	log.Debug("HandleMessageWrapper: to readMsg", "peer", peer)
	msg, err := peer.RW().ReadMsg()
	log.Debug("HandleMessageWrapper: recved msg", "peer", peer, "size", msg.Size, "e", err)
	if err != nil {
		return err
	}
	if msg.Size > ProtocolMaxMsgSize {
		return ErrMsgTooLarge
	}
	defer msg.Discard()

	data := &PttData{}
	err = msg.Decode(data)
	if err != nil {
		return nil
	}

	err = p.HandleMessage(CodeType(msg.Code), data, peer)
	if err != nil {
		log.Error("HandleMessageWrapper: unable to handle-msg", "code", msg.Code, "e", err, "peer", peer)
		return err
	}

	return nil
}

func (p *BasePtt) HandleMessage(code CodeType, data *PttData, peer *PttPeer) error {
	var err error

	//log.Debug("HandleMessage: start", "code", code, "data", data, "peer", peer)

	if !reflect.DeepEqual(data.Node, discover.EmptyNodeID) && !reflect.DeepEqual(data.Node, p.myNodeID[:]) {
		return ErrInvalidData
	}

	evCode, evHash, _, err := p.UnmarshalData(data)
	if err != nil {
		log.Error("HandleMessage: unable to unmarshal", "data", data, "e", err)
		return err
	}

	if evCode != code || !reflect.DeepEqual(evHash[:], data.Hash[:]) {
		log.Error("HandleMessage: hash not match", "evHash", evHash, "dataHash", data.Hash)
		return ErrInvalidData
	}

	if err != nil {
		log.Error("Ptt.HandleMessage", "code", code, "e", err)
	}

	return nil
}
