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

package types

import (
	"encoding/binary"
	"time"

	"github.com/ailabstw/go-pttai/common"
)

type Timestamp struct {
	Ts     int64  `json:"T"` // uint64 is only for
	NanoTs uint32 `json:"NT"`
}

var GetTimestamp = func() (Timestamp, error) {
	now := time.Now().UTC()

	return TimeToTimestamp(now), nil
}

func TimeToTimestamp(t time.Time) Timestamp {
	return Timestamp{int64(t.Unix()), uint32(t.Nanosecond())}
}

func (t *Timestamp) ToMilli() Timestamp {
	return Timestamp{t.Ts, (t.NanoTs / common.MILLION) * common.MILLION}
}

func (t *Timestamp) ToHRTimestamp() (Timestamp, Timestamp) {
	ts := t.Ts - t.Ts%HRSeconds
	return Timestamp{ts, 0}, Timestamp{ts + HRSeconds, 0}
}

func (t *Timestamp) ToDayTimestamp() (Timestamp, Timestamp) {
	ts := t.Ts - t.Ts%DaySeconds
	return Timestamp{ts, 0}, Timestamp{ts + DaySeconds, 0}
}

func (t *Timestamp) ToMonthTimestamp() (Timestamp, Timestamp) {
	theTime := time.Unix(t.Ts, int64(t.NanoTs)).UTC()
	year, month, _ := theTime.Date()

	newTime := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	newNextTime := newTime.AddDate(0, 1, 0)

	return TimeToTimestamp(newTime), TimeToTimestamp(newNextTime)
}

func (t *Timestamp) ToYearTimestamp() (Timestamp, Timestamp) {
	theTime := time.Unix(t.Ts, int64(t.NanoTs)).UTC()
	year, _, _ := theTime.Date()

	newTime := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	newNextTime := newTime.AddDate(1, 0, 0)

	return TimeToTimestamp(newTime), TimeToTimestamp(newNextTime)
}

func (t *Timestamp) IsEqMilli(t2 Timestamp) bool {
	return t.Ts == t2.Ts && t.NanoTs/common.MILLION == t2.NanoTs/common.MILLION
}

func (t *Timestamp) Marshal() ([]byte, error) {
	theBytes := make([]byte, 12) // uint64 + uint32
	binary.BigEndian.PutUint64(theBytes[:8], uint64(t.Ts))
	binary.BigEndian.PutUint32(theBytes[8:], t.NanoTs)

	return theBytes, nil
}

func UnmarshalTimestamp(theBytes []byte) (Timestamp, error) {
	ts := binary.BigEndian.Uint64(theBytes[:8])
	nanoTs := binary.BigEndian.Uint32(theBytes[8:])

	if nanoTs >= common.BILLION {
		return Timestamp{}, ErrInvalidTimestamp
	}

	return Timestamp{int64(ts), nanoTs}, nil
}

func (t *Timestamp) IsEqual(t2 Timestamp) bool {
	return t.Ts == t2.Ts && t.NanoTs == t2.NanoTs
}

func (t *Timestamp) IsLess(t2 Timestamp) bool {
	switch {
	case t.Ts < t2.Ts:
		return true
	case t.Ts > t2.Ts:
		return false
	}

	// equal ts
	switch {
	case t.NanoTs < t2.NanoTs:
		return true
	case t.NanoTs > t2.NanoTs:
		return false
	}

	return false
}

func (t *Timestamp) IsLessEqual(t2 Timestamp) bool {
	switch {
	case t.Ts < t2.Ts:
		return true
	case t.Ts > t2.Ts:
		return false
	}

	// equal ts
	switch {
	case t.NanoTs < t2.NanoTs:
		return true
	case t.NanoTs > t2.NanoTs:
		return false
	}

	return true
}

func MinTimestamp(a Timestamp, b Timestamp) Timestamp {
	if a.IsLess(b) {
		return a
	}

	return b
}
