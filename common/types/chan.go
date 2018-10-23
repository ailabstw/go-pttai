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
	"sync"

	"github.com/ailabstw/go-pttai/log"
)

/*
Chan implements non-panic chan.
Chan checks isClose first before passing to the channel.
*/
type Chan struct {
	lock      sync.Mutex
	lockClose sync.RWMutex
	theChan   chan interface{}
	isClose   bool
	n         int
}

/*
NewChan initialize the chan, the default of n should be 1.
*/
func NewChan(n int) *Chan {
	// defensive programming for n
	if n < 1 {
		n = 1
	}

	return &Chan{
		theChan: make(chan interface{}, n),
		n:       n,
	}
}

/*
Close closes the chan.

Instead of really closing the channel and may cause panic, we just setup the flag.
*/
func (c *Chan) Close() {
	// do close

	// early unlock c.to-close
	//c.lockClose.Lock()
	if c.IsClosed() {
		//c.lockClose.Unlock()
		return
	}
	c.isClose = true
	//c.lockClose.Unlock()

	log.Info("after check-close. to Lock")

	log.Info("after check-close. to close chan")
	c.Lock()
	defer c.Unlock()

	close(c.theChan)

	log.Info("after check-close. after close chan")
}

func (c *Chan) PassChan(d interface{}) error {

	//c.lockClose.RLock()
	if c.IsClosed() {
		//c.lockClose.RUnlock()
		return ErrLockClosed
	}
	//c.lockClose.RUnlock()

	c.Lock()
	defer c.Unlock()

	c.theChan <- d

	return nil
}

func (c *Chan) GetChan() chan interface{} {
	return c.theChan
}

func (c *Chan) IsClosed() bool {
	return c.isClose
}

func (c *Chan) IsBusy() bool {
	return len(c.theChan) == c.n
}

func (c *Chan) Lock() {
	c.lock.Lock()
}

func (c *Chan) Unlock() {
	c.lock.Unlock()
}
