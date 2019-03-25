// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/cheekybits/genny

package quic

import (
	"fmt"
	"sync"

	"github.com/pions/quic-go/internal/protocol"
	"github.com/pions/quic-go/internal/qerr"
	"github.com/pions/quic-go/internal/wire"
)

type outgoingUniStreamsMap struct {
	mutex sync.RWMutex
	cond  sync.Cond

	streams map[protocol.StreamID]sendStreamI

	nextStream   protocol.StreamID // stream ID of the stream returned by OpenStream(Sync)
	maxStream    protocol.StreamID // the maximum stream ID we're allowed to open
	maxStreamSet bool              // was maxStream set. If not, it's not possible to any stream (also works for stream 0)
	blockedSent  bool              // was a STREAMS_BLOCKED sent for the current maxStream

	newStream            func(protocol.StreamID) sendStreamI
	queueStreamIDBlocked func(*wire.StreamsBlockedFrame)

	closeErr error
}

func newOutgoingUniStreamsMap(
	nextStream protocol.StreamID,
	newStream func(protocol.StreamID) sendStreamI,
	queueControlFrame func(wire.Frame),
) *outgoingUniStreamsMap {
	m := &outgoingUniStreamsMap{
		streams:              make(map[protocol.StreamID]sendStreamI),
		nextStream:           nextStream,
		newStream:            newStream,
		queueStreamIDBlocked: func(f *wire.StreamsBlockedFrame) { queueControlFrame(f) },
	}
	m.cond.L = &m.mutex
	return m
}

func (m *outgoingUniStreamsMap) OpenStream() (sendStreamI, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	str, err := m.openStreamImpl()
	if err != nil {
		return nil, streamOpenErr{err}
	}
	return str, nil
}

func (m *outgoingUniStreamsMap) OpenStreamSync() (sendStreamI, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for {
		str, err := m.openStreamImpl()
		if err == nil {
			return str, nil
		}
		if err != nil && err != errTooManyOpenStreams {
			return nil, streamOpenErr{err}
		}
		m.cond.Wait()
	}
}

func (m *outgoingUniStreamsMap) openStreamImpl() (sendStreamI, error) {
	if m.closeErr != nil {
		return nil, m.closeErr
	}
	if !m.maxStreamSet || m.nextStream > m.maxStream {
		if !m.blockedSent {
			if m.maxStreamSet {
				m.queueStreamIDBlocked(&wire.StreamsBlockedFrame{
					Type:        protocol.StreamTypeUni,
					StreamLimit: m.maxStream.StreamNum(),
				})
			} else {
				m.queueStreamIDBlocked(&wire.StreamsBlockedFrame{
					Type:        protocol.StreamTypeUni,
					StreamLimit: 0,
				})
			}
			m.blockedSent = true
		}
		return nil, errTooManyOpenStreams
	}
	s := m.newStream(m.nextStream)
	m.streams[m.nextStream] = s
	m.nextStream += 4
	return s, nil
}

func (m *outgoingUniStreamsMap) GetStream(id protocol.StreamID) (sendStreamI, error) {
	m.mutex.RLock()
	if id >= m.nextStream {
		m.mutex.RUnlock()
		return nil, qerr.Error(qerr.InvalidStreamID, fmt.Sprintf("peer attempted to open stream %d", id))
	}
	s := m.streams[id]
	m.mutex.RUnlock()
	return s, nil
}

func (m *outgoingUniStreamsMap) DeleteStream(id protocol.StreamID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.streams[id]; !ok {
		return fmt.Errorf("Tried to delete unknown stream %d", id)
	}
	delete(m.streams, id)
	return nil
}

func (m *outgoingUniStreamsMap) SetMaxStream(id protocol.StreamID) {
	m.mutex.Lock()
	if !m.maxStreamSet || id > m.maxStream {
		m.maxStream = id
		m.maxStreamSet = true
		m.blockedSent = false
		m.cond.Broadcast()
	}
	m.mutex.Unlock()
}

func (m *outgoingUniStreamsMap) CloseWithError(err error) {
	m.mutex.Lock()
	m.closeErr = err
	for _, str := range m.streams {
		str.closeForShutdown(err)
	}
	m.cond.Broadcast()
	m.mutex.Unlock()
}
