// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package discover

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ailabstw/go-pttai/key"
	"github.com/ailabstw/go-pttai/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

const NodeIDBits = 512

// Node represents a host on the network.
// The fields of Node may not be modified.
type Node struct {
	IP       net.IP // len 4 for IPv4 or 16 for IPv6
	UDP, TCP uint16 // port numbers
	ID       NodeID // the node's public key

	IsP2P    bool
	PeerID   peer.ID
	PeerInfo *pstore.PeerInfo `rlp:"-"`

	// This is a cached copy of sha3(ID) which is used for node
	// distance calculations. This is part of Node in order to make it
	// possible to write tests that need a node at a certain distance.
	// In those tests, the content of sha will not actually correspond
	// with ID.
	sha common.Hash

	// Time when the node was added to the table.
	addedAt time.Time
}

// NewNode creates a new node. It is mostly meant to be used for
// testing purposes.
func NewNode(id NodeID, ip net.IP, udpPort, tcpPort uint16) *Node {
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	return &Node{
		IP:  ip,
		UDP: udpPort,
		TCP: tcpPort,
		ID:  id,
		sha: crypto.Keccak256Hash(id[:]),
	}
}

func NewP2PNode(id NodeID, peerID peer.ID, peerInfo *pstore.PeerInfo) *Node {
	return &Node{
		ID:       id,
		IsP2P:    true,
		PeerID:   peerID,
		PeerInfo: peerInfo,
		sha:      crypto.Keccak256Hash(id[:]),
	}
}

func NewP2PNodeWithNodeID(id NodeID) (*Node, error) {
	peerID, err := NodeIDToPeerID(id)
	if err != nil {
		return nil, err
	}
	return &Node{
		ID:     id,
		IsP2P:  true,
		PeerID: peerID,
		sha:    crypto.Keccak256Hash(id[:]),
	}, nil
}

func NewMyPeerInfo(peerID peer.ID, ip net.IP, tcpPort uint64) (*pstore.PeerInfo, error) {
	m, err := NewMyMultiaddr(ip, tcpPort)
	if err != nil {
		return nil, err
	}

	return &pstore.PeerInfo{
		ID:    peerID,
		Addrs: []ma.Multiaddr{m},
	}, nil
}

func NewMyMultiaddr(ip net.IP, tcpPort uint64) (ma.Multiaddr, error) {
	tcpStr := strconv.Itoa(int(tcpPort))
	maStr := "/ip4/" + ip.String() + "/tcp/" + tcpStr

	m, err := ma.NewMultiaddr(maStr)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (n *Node) addr() *net.UDPAddr {
	return &net.UDPAddr{IP: n.IP, Port: int(n.UDP)}
}

// Incomplete returns true for nodes with no IP address.
func (n *Node) Incomplete() bool {
	if n.IsP2P {
		return n.PeerInfo == nil
	}

	return n.IP == nil
}

// checks whether n is a valid complete node.
func (n *Node) validateComplete() error {
	if n.Incomplete() {
		return errors.New("incomplete node")
	}

	if !n.IsP2P {
		if n.UDP == 0 {
			return errors.New("missing UDP port")
		}
		if n.TCP == 0 {
			return errors.New("missing TCP port")
		}
		if n.IP.IsMulticast() || n.IP.IsUnspecified() {
			return errors.New("invalid IP (multicast/unspecified)")
		}
	}

	_, err := n.ID.Pubkey() // validate the key (on curve, etc.)
	return err
}

// The string representation of a Node is a URL.
// Please see ParseNode for a description of the format.
func (n *Node) String() string {
	if n.IsP2P {
		return n.p2pString()
	}

	u := url.URL{Scheme: PNode}
	if n.Incomplete() {
		u.Host = fmt.Sprintf("%x", n.ID[:])
	} else {
		u.User = url.User(fmt.Sprintf("%x", n.ID[:]))
		addr := net.TCPAddr{IP: n.IP, Port: int(n.TCP)}
		u.Host = addr.String()
		if n.UDP != n.TCP {
			u.RawQuery = "discport=" + strconv.Itoa(int(n.UDP))
		}
	}
	return u.String()
}

func (n *Node) p2pString() string {
	u := url.URL{Scheme: PNode}

	u.Host = peer.IDB58Encode(n.PeerID)
	u.RawQuery = "p2p=1"

	return u.String()
}

var incompleteNodeURL = regexp.MustCompile("(?i)^(?:pnode://)?([0-9a-f]+)$")
var incompleteP2PNodeURL = regexp.MustCompile("(?i)^(?:pnode://)?([0-9A-Za-z]+)$")

// ParseNode parses a node designator.
//
// There are two basic forms of node designators
//   - incomplete nodes, which only have the public key (node ID)
//   - complete nodes, which contain the public key and IP/Port information
//
// For incomplete nodes, the designator must look like one of these
//
//    pnode://<hex node id>
//    <hex node id>
//
// For complete nodes, the node ID is encoded in the username portion
// of the URL, separated from the host by an @ sign. The hostname can
// only be given as an IP address, DNS domain names are not allowed.
// The port in the host name section is the TCP listening port. If the
// TCP and UDP (discovery) ports differ, the UDP port is specified as
// query parameter "discport".
//
// In the following example, the node URL describes
// a node with IP address 10.3.58.6, TCP listening port 30303
// and UDP discovery port 30301.
//
//    pnode://<hex node id>@10.3.58.6:30303?discport=30301
func ParseP2PNode(rawurl string) (*Node, error) {
	if m := incompleteP2PNodeURL.FindStringSubmatch(rawurl); m != nil {
		log.Debug("ParseP2PNode: incomplete P2P Node", "m", m[1])
		peerID, err := peer.IDB58Decode(m[1])
		if err != nil {
			return nil, ErrInvalidURL
		}
		id, err := PeerIDToNodeID(peerID)
		if err != nil {
			return nil, ErrInvalidURL
		}
		return NewP2PNode(id, peerID, nil), nil
	}
	return parseP2PComplete(rawurl)
}

func parseP2PComplete(rawurl string) (*Node, error) {
	var (
		id      NodeID
		peerID  peer.ID
		ip      net.IP
		tcpPort uint64
	)

	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if u.Scheme != PNode {
		return nil, ErrInvalidURL
	}
	// Parse the Node ID from the user portion.
	if u.User == nil {
		return nil, ErrInvalidURL
	}

	peerID, err = peer.IDB58Decode(u.User.String())
	log.Debug("parseP2Complete: after ID", "peerID", peerID, "len", len(peerID), "e", err)
	if err != nil {
		return nil, ErrInvalidURL
	}

	// Parse the IP address.
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, ErrInvalidURL
	}
	if ip = net.ParseIP(host); ip == nil {
		return nil, ErrInvalidURL
	}
	// Ensure the IP is 4 bytes long for IPv4 addresses.
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	// Parse the port numbers.
	if tcpPort, err = strconv.ParseUint(port, 10, 16); err != nil {
		return nil, ErrInvalidURL
	}

	id, err = PeerIDToNodeID(peerID)
	if err != nil {
		return nil, err
	}

	peerInfo, err := NewMyPeerInfo(peerID, ip, tcpPort)
	if err != nil {
		return nil, err
	}

	return NewP2PNode(id, peerID, peerInfo), nil
}

// MustParseP2PNode parses a p2pnode URL. It panics if the URL is not valid.
func MustParseP2PNode(rawurl string) *Node {
	n, err := ParseP2PNode(rawurl)
	if err != nil {
		panic("invalid node URL: " + err.Error())
	}
	return n
}

func ParseNode(rawurl string) (*Node, error) {
	if m := incompleteNodeURL.FindStringSubmatch(rawurl); m != nil {
		id, err := HexID(m[1])
		if err != nil {
			return nil, fmt.Errorf("invalid node ID (%v)", err)
		}
		return NewNode(id, nil, 0, 0), nil
	}
	return parseComplete(rawurl)
}

func parseComplete(rawurl string) (*Node, error) {
	var (
		id               NodeID
		ip               net.IP
		tcpPort, udpPort uint64
	)
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "pnode" {
		return nil, errors.New("invalid URL scheme, want \"pnode\"")
	}
	// Parse the Node ID from the user portion.
	if u.User == nil {
		return nil, errors.New("does not contain node ID")
	}
	if id, err = HexID(u.User.String()); err != nil {
		return nil, fmt.Errorf("invalid node ID (%v)", err)
	}
	// Parse the IP address.
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host: %v", err)
	}
	if ip = net.ParseIP(host); ip == nil {
		return nil, errors.New("invalid IP address")
	}
	// Ensure the IP is 4 bytes long for IPv4 addresses.
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	// Parse the port numbers.
	if tcpPort, err = strconv.ParseUint(port, 10, 16); err != nil {
		return nil, errors.New("invalid port")
	}
	udpPort = tcpPort
	qv := u.Query()
	if qv.Get("discport") != "" {
		udpPort, err = strconv.ParseUint(qv.Get("discport"), 10, 16)
		if err != nil {
			return nil, errors.New("invalid discport in query")
		}
	}
	return NewNode(id, ip, uint16(udpPort), uint16(tcpPort)), nil
}

// MustParseNode parses a node URL. It panics if the URL is not valid.
func MustParseNode(rawurl string) *Node {
	n, err := ParseNode(rawurl)
	if err != nil {
		panic("invalid node URL: " + err.Error())
	}
	return n
}

// MarshalText implements encoding.TextMarshaler.
func (n *Node) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *Node) UnmarshalText(text []byte) error {
	dec, err := ParseNode(string(text))
	if err == nil {
		*n = *dec
	}
	return err
}

// NodeID is a unique identifier for each node.
// The node identifier is a marshaled elliptic curve public key.
type NodeID [NodeIDBits / 8]byte

// Bytes returns a byte slice representation of the NodeID
func (n NodeID) Bytes() []byte {
	return n[:]
}

// NodeID prints as a long hexadecimal number.
func (n NodeID) String() string {
	return fmt.Sprintf("%x", n[:])
}

// The Go syntax representation of a NodeID is a call to HexID.
func (n NodeID) GoString() string {
	return fmt.Sprintf("discover.HexID(\"%x\")", n[:])
}

// TerminalString returns a shortened hex string for terminal logging.
func (n NodeID) TerminalString() string {
	return hex.EncodeToString(n[:8])
}

// MarshalText implements the encoding.TextMarshaler interface.
func (n NodeID) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(n[:])), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (n *NodeID) UnmarshalText(text []byte) error {
	id, err := HexID(string(text))
	if err != nil {
		return err
	}
	*n = id
	return nil
}

func (n *NodeID) ToRaftID() (uint64, error) {
	pubkey, err := n.Pubkey()
	if err != nil {
		return 0, err
	}

	addr := crypto.PubkeyToAddress(*pubkey)
	id := binary.BigEndian.Uint64(addr[OffsetAddrToRaftID:OffsetEndAddrToRaftID])

	if id == 0 {
		return 0, ErrInvalidNodeID
	}

	return id, nil
}

// BytesID converts a byte slice to a NodeID
func BytesID(b []byte) (NodeID, error) {
	var id NodeID
	if len(b) != len(id) {
		return id, fmt.Errorf("wrong length, want %d bytes", len(id))
	}
	copy(id[:], b)
	return id, nil
}

// MustBytesID converts a byte slice to a NodeID.
// It panics if the byte slice is not a valid NodeID.
func MustBytesID(b []byte) NodeID {
	id, err := BytesID(b)
	if err != nil {
		panic(err)
	}
	return id
}

// HexID converts a hex string to a NodeID.
// The string may be prefixed with 0x.
func HexID(in string) (NodeID, error) {
	var id NodeID
	b, err := hex.DecodeString(strings.TrimPrefix(in, "0x"))
	if err != nil {
		return id, err
	} else if len(b) != len(id) {
		return id, fmt.Errorf("wrong length, want %d hex chars", len(id)*2)
	}
	copy(id[:], b)
	return id, nil
}

// MustHexID converts a hex string to a NodeID.
// It panics if the string is not a valid NodeID.
func MustHexID(in string) NodeID {
	id, err := HexID(in)
	if err != nil {
		panic(err)
	}
	return id
}

// PubkeyID returns a marshaled representation of the given public key.
func PubkeyID(pub *ecdsa.PublicKey) NodeID {
	var id NodeID
	pbytes := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	if len(pbytes)-1 != len(id) {
		panic(fmt.Errorf("need %d bit pubkey, got %d bits", (len(id)+1)*8, len(pbytes)))
	}
	copy(id[:], pbytes[1:])
	return id
}

// Pubkey returns the public key represented by the node ID.
// It returns an error if the ID is not a point on the curve.
func (id NodeID) Pubkey() (*ecdsa.PublicKey, error) {
	p := &ecdsa.PublicKey{Curve: crypto.S256(), X: new(big.Int), Y: new(big.Int)}
	half := len(id) / 2
	p.X.SetBytes(id[:half])
	p.Y.SetBytes(id[half:])
	if !p.Curve.IsOnCurve(p.X, p.Y) {
		return nil, errors.New("id is invalid secp256k1 curve point")
	}
	return p, nil
}

// recoverNodeID computes the public key used to sign the
// given hash from the signature.
func recoverNodeID(hash, sig []byte) (id NodeID, err error) {
	pubkey, err := secp256k1.RecoverPubkey(hash, sig)
	if err != nil {
		return id, err
	}
	if len(pubkey)-1 != len(id) {
		return id, fmt.Errorf("recovered pubkey has %d bits, want %d bits", len(pubkey)*8, (len(id)+1)*8)
	}
	for i := range id {
		id[i] = pubkey[i+1]
	}
	return id, nil
}

// distcmp compares the distances a->target and b->target.
// Returns -1 if a is closer to target, 1 if b is closer to target
// and 0 if they are equal.
func distcmp(target, a, b common.Hash) int {
	for i := range target {
		da := a[i] ^ target[i]
		db := b[i] ^ target[i]
		if da > db {
			return 1
		} else if da < db {
			return -1
		}
	}
	return 0
}

// table of leading zero counts for bytes [0..255]
var lzcount = [256]int{
	8, 7, 6, 6, 5, 5, 5, 5,
	4, 4, 4, 4, 4, 4, 4, 4,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

// logdist returns the logarithmic distance between a and b, log2(a ^ b).
func logdist(a, b common.Hash) int {
	lz := 0
	for i := range a {
		x := a[i] ^ b[i]
		if x == 0 {
			lz += 8
		} else {
			lz += lzcount[x]
			break
		}
	}
	return len(a)*8 - lz
}

// hashAtDistance returns a random hash such that logdist(a, b) == n
func hashAtDistance(a common.Hash, n int) (b common.Hash) {
	if n == 0 {
		return a
	}
	// flip bit at position n, fill the rest with random bits
	b = a
	pos := len(a) - n/8 - 1
	bit := byte(0x01) << (byte(n%8) - 1)
	if bit == 0 {
		pos++
		bit = 0x80
	}
	b[pos] = a[pos]&^bit | ^a[pos]&bit // TODO: randomize end bits
	for i := pos + 1; i < len(a); i++ {
		b[i] = byte(rand.Intn(255))
	}
	return b
}

func GenerateNodeKey() (*ecdsa.PrivateKey, error) {
	for i := 0; i < NGenerateNodeKey; i++ {
		key, err := key.GenerateKey()
		if err != nil {
			continue
		}

		nodeID := PubkeyID(&key.PublicKey)
		_, err = nodeID.ToRaftID()
		if err != nil {
			continue
		}

		return key, nil
	}

	return nil, ErrInvalidKey
}

func LoadECDSA(filename string) (*ecdsa.PrivateKey, error) {
	key, err := crypto.LoadECDSA(filename)
	if err != nil {
		return nil, err
	}

	nodeID := PubkeyID(&key.PublicKey)
	_, err = nodeID.ToRaftID()
	if err != nil {
		return nil, err
	}

	return key, nil
}

func PeerIDToNodeID(peerID peer.ID) (NodeID, error) {
	if len(peerID) == LenPeerIDNoPubkey {
		return NodeID{}, nil
	}

	p2pPubKey, err := peerID.ExtractPublicKey()
	if err != nil {
		return NodeID{}, err
	}

	pubKey, err := key.P2PPubkeyToPubkey(p2pPubKey)
	if err != nil {
		return NodeID{}, err
	}

	return PubkeyID(pubKey), nil
}

func NodeIDToPeerID(nodeID NodeID) (peer.ID, error) {
	pubKey, err := nodeID.Pubkey()
	if err != nil {
		return "", err
	}

	p2pPubKey, err := key.PubKeyToP2PPubkey(pubKey)
	if err != nil {
		return "", err
	}

	return peer.IDFromPublicKey(p2pPubKey)
}
