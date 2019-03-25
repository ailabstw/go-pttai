package quic

import (
	"fmt"

	"github.com/pions/quic-go/internal/protocol"
	"github.com/pions/quic-go/internal/utils"
)

type serverSession struct {
	quicSession

	config *Config

	logger utils.Logger
}

var _ packetHandler = &serverSession{}

func newServerSession(sess quicSession, config *Config, logger utils.Logger) packetHandler {
	return &serverSession{
		quicSession: sess,
		config:      config,
		logger:      logger,
	}
}

func (s *serverSession) handlePacket(p *receivedPacket) {
	if err := s.handlePacketImpl(p); err != nil {
		s.logger.Debugf("error handling packet from %s: %s", p.remoteAddr, err)
	}
}

func (s *serverSession) handlePacketImpl(p *receivedPacket) error {
	hdr := p.hdr

	// Probably an old packet that was sent by the client before the version was negotiated.
	// It is safe to drop it.
	if hdr.IsLongHeader && hdr.Version != s.quicSession.GetVersion() {
		return nil
	}

	if hdr.IsLongHeader {
		switch hdr.Type {
		case protocol.PacketTypeInitial, protocol.PacketTypeHandshake:
			// nothing to do here. Packet will be passed to the session.
		default:
			// Note that this also drops 0-RTT packets.
			return fmt.Errorf("Received unsupported packet type: %s", hdr.Type)
		}
	}

	s.quicSession.handlePacket(p)
	return nil
}

func (s *serverSession) GetPerspective() protocol.Perspective {
	return protocol.PerspectiveServer
}
