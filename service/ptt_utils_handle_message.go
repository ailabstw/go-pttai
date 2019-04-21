// Copyright 2019 The go-pttai Authors
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
	"github.com/ethereum/go-ethereum/common"
)

/*
HandleMessageWrapper
*/
func (p *BasePtt) HandleMessageWrapper(peer *PttPeer) error {
	log.Debug("HandleMessageWrapper: to ReadMsg", "peer", peer)
	msg, err := peer.RW().ReadMsg()
	log.Debug("HandleMessageWrapper: after ReadMsg", "e", err, "code", msg.Code, "size", msg.Size)
	if err != nil {
		log.Error("HandleMessageWrapper: unable ReadMsg", "peer", peer, "e", err)
		return err
	}
	defer msg.Discard()

	if msg.Size > ProtocolMaxMsgSize {
		log.Error("HandleMessageWrapper: exceed size", "peer", peer, "msg.Size", msg.Size)
		return ErrMsgTooLarge
	}

	data := &PttData{}
	err = msg.Decode(data)
	if err != nil {
		log.Error("HandleMessageWrapper: unable to decode data", "peer", peer, "e", err)
		return err
	}

	err = p.HandleMessage(CodeType(msg.Code), data, peer)
	if err != nil {
		log.Error("HandleMessageWrapper: unable to handle-msg", "code", msg.Code, "e", err, "peer", peer)
		return err
	}

	return nil
}

/*
HandleMessage handles message
*/
func (p *BasePtt) HandleMessage(code CodeType, data *PttData, peer *PttPeer) error {
	var err error

	log.Debug("HandleMessage: start", "code", code, "peer", peer, "peerType", peer.PeerType)

	if !reflect.DeepEqual(data.Node, discover.EmptyNodeID) && !reflect.DeepEqual(data.Node, p.myNodeID[:]) {
		log.Error("HandleMessage: the msg is not for me or not for broadcast", "code", code, "data.Node", data.Node, "peer", peer)
		return ErrInvalidData
	}

	evCode, evHash, encData, err := p.UnmarshalData(data)
	if err != nil {
		log.Error("HandleMessage: unable to unmarshal", "data", data, "e", err)
		return err
	}

	if evCode != code || (code < CodeTypeRequireHash && !reflect.DeepEqual(evHash[:], data.Hash[:])) {
		log.Error("HandleMessage: hash not match", "evHash", evHash, "dataHash", data.Hash)
		return ErrInvalidData
	}

	switch code {
	case CodeTypeJoin:
		err = p.HandleCodeJoin(evHash, encData, peer)
	case CodeTypeJoinAck:
		err = p.HandleCodeJoinAck(evHash, encData, peer)

	case CodeTypeOp:
		err = p.HandleCodeOp(evHash, encData, peer)
	case CodeTypeOpFail:
		err = p.HandleCodeOpFail(evHash, encData, peer)

	case CodeTypeRequestOpKey:
		err = p.HandleCodeRequestOpKey(evHash, encData, peer)
	case CodeTypeRequestOpKeyFail:
		err = p.HandleCodeRequestOpKeyFail(evHash, encData, peer)
	case CodeTypeRequestOpKeyAck:
		err = p.HandleCodeRequestOpKeyAck(evHash, encData, peer)

	case CodeTypeEntityDeleted:
		err = p.HandleCodeEntityDeleted(evHash, encData, peer)

	case CodeTypeOpCheckMember:
		err = p.HandleCodeOpCheckMember(evHash, encData, peer)
	case CodeTypeOpCheckMemberAck:
		err = p.HandleCodeOpCheckMemberAck(evHash, encData, peer)

	case CodeTypeIdentifyPeer:
		err = p.HandleCodeIdentifyPeer(evHash, encData, peer)
	case CodeTypeIdentifyPeerFail:
		err = p.HandleCodeIdentifyPeerFail(evHash, encData, peer)
	case CodeTypeIdentifyPeerWithMyID:
		err = p.HandleCodeIdentifyPeerWithMyID(evHash, encData, peer)
	case CodeTypeIdentifyPeerWithMyIDChallenge:
		err = p.HandleCodeIdentifyPeerWithMyIDChallenge(evHash, encData, peer)
	case CodeTypeIdentifyPeerWithMyIDChallengeAck:
		err = p.HandleCodeIdentifyPeerWithMyIDChallengeAck(evHash, encData, peer)
	case CodeTypeIdentifyPeerWithMyIDAck:
		err = p.HandleCodeIdentifyPeerWithMyIDAck(evHash, encData, peer)
	default:
		err = ErrInvalidMsgCode
	}

	if err != nil {
		log.Error("Ptt.HandleMessage", "code", code, "e", err)

	}

	return nil
}

func (p *BasePtt) HandleCodeJoin(hash *common.Address, encData []byte, peer *PttPeer) error {
	entity, err := p.getEntityFromHash(hash, &p.lockJoins, p.joins)
	if err != nil {
		log.Error("HandleCodeJoin: getEntityFromHash", "e", err)
		return err
	}

	pm := entity.PM()
	keyInfo, err := pm.GetJoinKeyFromHash(hash)
	if err != nil {
		log.Error("HandleCodeJoin: unable to get JoinKeyInfo", "hash", hash, "e", err)
		return err
	}

	op, dataBytes, err := p.DecryptData(encData, keyInfo)
	if err != nil {
		log.Error("HandleCodeJoin: unable to DecryptData", "e", err)
		return err
	}

	log.Debug("HandleCodeJoin: start", "op", op, "joinMsg", JoinMsg, "joinEntityMsg", JoinEntityMsg)

	switch op {
	case JoinMsg:
		err = p.HandleJoin(dataBytes, hash, entity, pm, keyInfo, peer)
	case JoinEntityMsg:
		err = p.HandleJoinEntity(dataBytes, hash, entity, pm, keyInfo, peer)
	default:
		err = ErrInvalidMsgCode
	}

	return err
}

func (p *BasePtt) HandleCodeJoinAck(hash *common.Address, encData []byte, peer *PttPeer) error {

	joinRequest, err := p.myEntity.GetJoinRequest(hash)
	log.Debug("HandleCodeJoinAck: after GetJoinRequest", "e", err)
	if err != nil {
		return err
	}

	keyInfo := joinKeyToKeyInfo(joinRequest.Key)

	op, dataBytes, err := p.DecryptData(encData, keyInfo)
	if err != nil {
		return err
	}

	log.Debug("HandleCodeJoinAck: start", "op", op, "ApproveJoinMsg", ApproveJoinMsg)

	switch op {
	case JoinAckChallengeMsg:
		err = p.HandleJoinAckChallenge(dataBytes, hash, joinRequest, peer)
	case ApproveJoinMsg:
		err = p.HandleApproveJoin(dataBytes, hash, joinRequest, peer)
	default:
		err = ErrInvalidMsgCode
	}

	return err
}

func (p *BasePtt) HandleCodeOp(hash *common.Address, encData []byte, peer *PttPeer) error {

	entity, err := p.getEntityFromHash(hash, &p.lockOps, p.ops)
	log.Debug("HandleCodeOp: after getEntityFromHash", "e", err, "hash", hash)
	if err != nil {
		log.Error("HandleCodeOp: invalid entity", "hash", hash, "e", err)
		return p.OpFail(hash, peer)
	}

	pm := entity.PM()

	err = PMHandleMessageWrapper(pm, hash, encData, peer)

	return err
}

func (p *BasePtt) HandleCodeOpFail(hash *common.Address, encData []byte, peer *PttPeer) error {

	return p.HandleOpFail(encData, peer)
}

func (p *BasePtt) HandleCodeRequestOpKey(hash *common.Address, encData []byte, peer *PttPeer) error {

	return p.HandleRequestOpKey(encData, peer)
}

func (p *BasePtt) HandleCodeRequestOpKeyFail(hash *common.Address, encData []byte, peer *PttPeer) error {

	return p.HandleRequestOpKeyFail(encData, peer)
}

func (p *BasePtt) HandleCodeRequestOpKeyAck(hash *common.Address, encData []byte, peer *PttPeer) error {
	return p.HandleRequestOpKeyAck(encData, peer)
}

func (p *BasePtt) HandleCodeEntityDeleted(hash *common.Address, encData []byte, peer *PttPeer) error {
	return p.HandleEntityTerminal(encData, peer)
}

func (p *BasePtt) HandleCodeOpCheckMember(hash *common.Address, encData []byte, peer *PttPeer) error {
	return p.HandleOpCheckMember(encData, peer)
}

func (p *BasePtt) HandleCodeOpCheckMemberAck(hash *common.Address, encData []byte, peer *PttPeer) error {
	return p.HandleOpCheckMemberAck(encData, peer)
}

func (p *BasePtt) HandleCodeIdentifyPeer(hash *common.Address, encData []byte, peer *PttPeer) error {

	entity, err := p.getEntityFromHash(hash, &p.lockOps, p.ops)
	if err != nil {
		log.Error("HandleCodeIdentifyPeer: invalid entity", "hash", hash, "e", err)
		return p.IdentifyPeerFail(hash, peer)
	}

	pm := entity.PM()

	err = PMHandleMessageWrapper(pm, hash, encData, peer)
	if err != nil {
		p.IdentifyPeerFail(hash, peer)
	}

	return err
}

func (p *BasePtt) HandleCodeIdentifyPeerFail(hash *common.Address, encData []byte, peer *PttPeer) error {

	return p.HandleIdentifyPeerFail(encData, peer)
}

func (p *BasePtt) HandleCodeIdentifyPeerWithMyID(hash *common.Address, encData []byte, peer *PttPeer) error {

	return p.HandleIdentifyPeerWithMyID(encData, peer)
}

func (p *BasePtt) HandleCodeIdentifyPeerWithMyIDChallenge(hash *common.Address, encData []byte, peer *PttPeer) error {

	return p.HandleIdentifyPeerWithMyIDChallenge(encData, peer)
}

func (p *BasePtt) HandleCodeIdentifyPeerWithMyIDChallengeAck(hash *common.Address, encData []byte, peer *PttPeer) error {

	return p.HandleIdentifyPeerWithMyIDChallengeAck(encData, peer)
}

func (p *BasePtt) HandleCodeIdentifyPeerWithMyIDAck(hash *common.Address, encData []byte, peer *PttPeer) error {

	return p.HandleIdentifyPeerWithMyIDAck(encData, peer)
}
