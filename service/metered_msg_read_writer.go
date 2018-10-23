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

import "github.com/ailabstw/go-pttai/p2p"

type MeteredMsgReadWriter interface {
	p2p.MsgReadWriter

	Version() uint
	Init(version uint) error
}

type BaseMeteredMsgReadWriter struct {
	p2p.MsgReadWriter

	version uint
}

func NewBaseMeteredMsgReadWriter(rw p2p.MsgReadWriter, version uint) (MeteredMsgReadWriter, error) {
	return &BaseMeteredMsgReadWriter{
		MsgReadWriter: rw,
	}, nil
}

func (rw *BaseMeteredMsgReadWriter) Version() uint {
	return rw.version
}

func (rw *BaseMeteredMsgReadWriter) Init(version uint) error {
	rw.version = version

	return nil
}

func (rw *BaseMeteredMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	msg, err := rw.MsgReadWriter.ReadMsg()
	if err != nil {
		return msg, err
	}

	return msg, nil
}

func (rw *BaseMeteredMsgReadWriter) WriteMsg(msg p2p.Msg) error {
	return rw.MsgReadWriter.WriteMsg(msg)
}
