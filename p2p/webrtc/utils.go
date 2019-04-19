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

package webrtc

import (
	"strings"

	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/pion/webrtc"
)

// https://godoc.org/github.com/pion/sdp#SessionDescription

// https://godoc.org/github.com/pion/sdp#SessionDescription
func parseOfferID(offer *webrtc.SessionDescription) (string, error) {
	desc := offer.SDP
	descList := strings.Split(desc, "\r\n")
	log.Debug("parseOfferID: to for-loop", "descList", descList)
	for _, eachDesc := range descList {
		log.Debug("parseOfferID (in-for-loop)", "eachDesc", eachDesc, "OfferIDPrefix", OfferIDPrefix)
		if strings.HasPrefix(eachDesc, OfferIDPrefix) {
			return eachDesc[OfferIDOffset:], nil
		}
	}

	return "", ErrInvalidWebrtcOffer
}

func parseWebrtcAddr(nodeID discv5.NodeID) *webrtcAddr {

	return &webrtcAddr{addr: nodeID.String()}
}

func dataChannelToWebrtcConn(
	dataChannel *webrtc.DataChannel,
	localID discv5.NodeID,
	fromID discv5.NodeID,
	peerConn *webrtc.PeerConnection,
) (*WebrtcConn, error) {

	dataConn, err := dataChannel.Detach()
	if err != nil {
		return nil, err
	}

	// XXX we may need the unified nodeID type.
	var nodeID discover.NodeID
	copy(nodeID[:], fromID[:])

	info := &webrtcInfo{
		NodeID: nodeID,

		PeerConn: peerConn,
		DataConn: dataConn,
	}

	conn, err := NewWebrtcConn(localID, fromID, info)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
