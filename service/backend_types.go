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
	"encoding/base64"
	"net"
	"net/url"
	"reflect"
	"strconv"

	"github.com/ailabstw/go-pttai/common"
	"github.com/ailabstw/go-pttai/common/types"
	"github.com/ailabstw/go-pttai/crypto"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
)

const (
	PathJoinMe     = "/joinme"
	PathJoinFriend = "/joinfriend"
	PathJoinBoard  = "/joinboard"
)

type BackendCountPeers struct {
	MyPeers        int `json:"M"`
	ImportantPeers int `json:"I"`
	MemberPeers    int `json:"E"`
	RandomPeers    int `json:"R"`
}

type BackendPeer struct {
	NodeID   *discover.NodeID `json:"ID"`
	PeerType PeerType         `json:"T"`
	UserID   *types.PttID     `json:"UID"`
	Addr     net.Addr
}

func PeerToBackendPeer(peer *PttPeer) *BackendPeer {
	return &BackendPeer{
		NodeID:   peer.GetID(),
		PeerType: peer.PeerType,
		UserID:   peer.UserID,
		Addr:     peer.RemoteAddr(),
	}
}

type BackendJoinURL struct {
	CreatorID    string `json:"C"`
	Name         string `json:"N"`
	Hash         string `json:"H"`
	Pn           string `json:"Pn"`
	URL          string
	UpdateTS     types.Timestamp `json:"UT"`
	ExpireSecond uint            `json:"e"`
}

type BackendJoinRequest struct {
	CreatorID *types.PttID     `json:"C"`
	NodeID    *discover.NodeID `json:"n"`
	Hash      []byte           `json:"H"`
	Name      []byte           `json:"N"`
	Status    JoinStatus       `json:"S"`
}

func MarshalBackendJoinURL(id *types.PttID, nodeID *discover.NodeID, keyInfo *KeyInfo, name []byte, path string) (*BackendJoinURL, error) {
	nodeIDBytes, err := nodeID.MarshalText()
	if err != nil {
		return nil, err
	}
	nodeIDStr := string(nodeIDBytes)

	creatorIDBytes, err := id.MarshalText()
	if err != nil {
		return nil, err
	}
	creatorIDStr := string(creatorIDBytes)

	keyBytes := crypto.FromECDSA(keyInfo.Key)
	keyStr := base64.StdEncoding.EncodeToString(keyBytes[:])

	keyHash := keyInfo.Hash
	keyHashStr := base64.StdEncoding.EncodeToString(keyHash[:])

	nameStr := base64.StdEncoding.EncodeToString(name)

	v := &url.Values{}
	v.Add("c", creatorIDStr)
	v.Add("h", keyHashStr)
	v.Add("k", keyStr)
	v.Add("n", nameStr)
	v.Add("t", strconv.FormatUint(keyInfo.UpdateTS.Ts+IntRenewJoinKeySeconds, 10))

	return &BackendJoinURL{
		CreatorID:    creatorIDStr,
		Name:         nameStr,
		Hash:         keyHashStr,
		Pn:           nodeIDStr,
		URL:          "pnode://" + nodeIDStr + path + "?" + v.Encode(),
		UpdateTS:     keyInfo.UpdateTS,
		ExpireSecond: IntRenewJoinKeySeconds,
	}, nil
}

func ParseBackendJoinURL(urlBytes []byte, path string) (*JoinRequest, error) {
	// parse url
	theURL, err := url.Parse(string(urlBytes))
	log.Debug("after parse", "urlBytes", string(urlBytes), "theURL", theURL, "Scheme", theURL.Scheme, "Path", theURL.Path, "Host", theURL.Host, "query", theURL.RawQuery, "e", err)
	if err != nil {
		return nil, types.ErrInvalidURL
	}

	if theURL.Scheme != "pnode" {
		return nil, types.ErrInvalidURL
	}

	if theURL.Path != path {
		return nil, types.ErrInvalidURL
	}

	// node
	nodeIDStr := theURL.Host
	if theURL.User != nil {
		nodeIDStr = theURL.User.Username()
	}

	log.Debug("to parse node", "host", theURL.Host, "user", theURL.User, "nodeIDstr", nodeIDStr)

	nodeIDBytes := []byte(nodeIDStr)

	nodeID := &discover.NodeID{}
	err = nodeID.UnmarshalText(nodeIDBytes)
	if err != nil {
		return nil, types.ErrInvalidURL
	}

	// parse query
	query, err := url.ParseQuery(theURL.RawQuery)
	log.Debug("after parse query", "RawQuery", theURL.RawQuery, "query", query, "e", err)

	if err != nil {
		return nil, types.ErrInvalidURL
	}

	// hash
	hashStr := query.Get("h")
	hashBytes, err := base64.StdEncoding.DecodeString(hashStr)
	if err != nil {
		return nil, types.ErrInvalidURL
	}
	hashVal := common.BytesToAddress(hashBytes)
	hash := &hashVal

	// key
	keyStr := query.Get("k")
	keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, types.ErrInvalidURL
	}

	key, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, types.ErrInvalidURL
	}

	log.Info("after parse query", "hash", hash, "key", key)

	keyHash := crypto.PubkeyToAddress(key.PublicKey)
	if !reflect.DeepEqual(hash, &keyHash) {
		return nil, types.ErrInvalidURL
	}

	// creator
	creatorIDStr := query.Get("c")
	creatorID := &types.PttID{}
	err = creatorID.UnmarshalText([]byte(creatorIDStr))
	log.Debug("after unmarshal creatorID", "creatorIDStr", creatorIDStr, "creatorID", creatorID, "e", err)
	if err != nil {
		return nil, types.ErrInvalidURL
	}

	// name
	nameStr := query.Get("n")
	name, err := base64.StdEncoding.DecodeString(nameStr)
	if err != nil {
		return nil, types.ErrInvalidURL
	}

	// ts
	ts, err := types.GetTimestamp()
	if err != nil {
		return nil, err
	}

	// challenge
	challenge := GenChallenge()

	return &JoinRequest{
		CreatorID: creatorID,
		CreateTS:  ts,
		NodeID:    nodeID,
		Hash:      hash,
		Key:       key,
		Name:      name,
		Status:    JoinStatusPending,
		Challenge: challenge,
	}, nil
}

func JoinRequestToBackendJoinRequest(joinRequest *JoinRequest) *BackendJoinRequest {
	return &BackendJoinRequest{
		CreatorID: joinRequest.CreatorID,
		NodeID:    joinRequest.NodeID,
		Hash:      joinRequest.Hash[:],
		Name:      joinRequest.Name,
		Status:    joinRequest.Status,
	}
}

type BackendConfirmJoin struct {
	ID         *types.PttID
	Name       []byte           `json:"N"`
	EntityID   *types.PttID     `json:"EID"`
	EntityName []byte           `json:"EN"`
	JoinType   JoinType         `json:"JT"`
	UpdateTS   types.Timestamp  `json:"UT"`
	NodeID     *discover.NodeID `json:"NID"`
}

type BackendMerkleNode MerkleNode

func MerkleNodeToBackendMerkleNode(m *MerkleNode) *BackendMerkleNode {
	return &BackendMerkleNode{
		Level:     m.Level,
		Addr:      m.Addr,
		UpdateTS:  m.UpdateTS,
		NChildren: m.NChildren,
		Key:       m.Key,
	}
}
