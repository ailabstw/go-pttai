package signalserver

import "github.com/ethereum/go-ethereum/p2p/discv5"

type Signal struct {
	FromID discv5.NodeID `json:"F,omitempty"`
	ToID   discv5.NodeID `json:"T,omitempty"`

	Msg   []byte `json:"M,omitempty"`
	Extra []byte `json:"E,omitempty"`
}
