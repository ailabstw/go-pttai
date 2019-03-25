package quic

import (
	"fmt"

	"github.com/pions/quic-go/internal/protocol"
	"github.com/pions/quic-go/internal/wire"
)

type cryptoDataHandler interface {
	HandleMessage([]byte, protocol.EncryptionLevel) bool
}

type cryptoStreamManager struct {
	cryptoHandler cryptoDataHandler

	initialStream   cryptoStream
	handshakeStream cryptoStream
}

func newCryptoStreamManager(
	cryptoHandler cryptoDataHandler,
	initialStream cryptoStream,
	handshakeStream cryptoStream,
) *cryptoStreamManager {
	return &cryptoStreamManager{
		cryptoHandler:   cryptoHandler,
		initialStream:   initialStream,
		handshakeStream: handshakeStream,
	}
}

func (m *cryptoStreamManager) HandleCryptoFrame(frame *wire.CryptoFrame, encLevel protocol.EncryptionLevel) (bool /* encryption level changed */, error) {
	var str cryptoStream
	switch encLevel {
	case protocol.EncryptionInitial:
		str = m.initialStream
	case protocol.EncryptionHandshake:
		str = m.handshakeStream
	default:
		return false, fmt.Errorf("received CRYPTO frame with unexpected encryption level: %s", encLevel)
	}
	if err := str.HandleCryptoFrame(frame); err != nil {
		return false, err
	}
	for {
		data := str.GetCryptoData()
		if data == nil {
			return false, nil
		}
		if encLevelFinished := m.cryptoHandler.HandleMessage(data, encLevel); encLevelFinished {
			return true, str.Finish()
		}
	}
}
