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

package raft

import (
	"context"
	"testing"
	"time"

	"github.com/ailabstw/go-pttai/log"
	"github.com/ailabstw/go-pttai/raft/raftpb"
	"github.com/stretchr/testify/assert"
)

func TestConfChangeAddNodeDefault(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	log.Debug("TestConfChangeAddNode: start")

	myAssert := assert.New(t)

	confFunc := func(c *Config) {
		c.PreVote = true
		c.peers = make(map[uint64]uint32)
		c.peers[c.ID] = uint32(1)
	}

	nt := newNetworkWithConfig(confFunc, nil, nil)
	log.Debug("after newNetworkWithConfig", "nt", nt)
	t.Logf("nt: %v", nt)

	n := newNode()
	s := nt.storage[1]
	//r := newTestRaft(1, []uint64{1}, 10, 1, s)
	r := nt.peers[1].(*raft)
	go n.run(r)

	n2 := newNode()
	s2 := nt.storage[2]
	//r2 := newTestRaft(2, []uint64{1}, 10, 1, s)
	r2 := nt.peers[2].(*raft)

	n.Campaign(context.TODO())
	rdyEntries := make([]raftpb.Entry, 0)

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	ticker2 := time.NewTicker(time.Millisecond * 10)
	defer ticker2.Stop()

	done := make(chan struct{})
	done2 := make(chan struct{})
	stop := make(chan struct{})
	applyConfChan := make(chan struct{})
	applyConfChan2 := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-stop: // stop
				return
			case <-ticker.C: // tick
				log.Debug("n.Tick: start")
				n.Tick()
				log.Debug("n.Tick: done")
			case rd := <-n.Ready(): // ready
				s.Append(rd.Entries)
				applied := false
				for _, e := range rd.Entries {
					rdyEntries = append(rdyEntries, e)
					switch e.Type {
					case raftpb.EntryNormal:
					case raftpb.EntryConfChange:
						var cc raftpb.ConfChange
						cc.Unmarshal(e.Data)
						n.ApplyConfChange(cc)
						applied = true
					}
				}
				log.Debug("n.Ready", "msgs", rd.Messages)
				for _, msg := range rd.Messages {
					nt.send(msg)
				}

				n.Advance()
				if applied {
					nodes, weights := r2.nodes(false)
					t.Logf("n.Ready: conf-changed: nodes: %v weights: %v", nodes, weights)
					applyConfChan <- struct{}{}
				}
			}
		}
	}()

	go func() {
		defer close(done2)
		for {
			select {
			case <-stop: // stop
				return
			case <-ticker2.C: // tick
				log.Debug("n2.Tick: start")
				n2.Tick()
				log.Debug("n2.Tick: done")
			case rd := <-n2.Ready(): // ready
				s2.Append(rd.Entries)
				applied := false
				for _, e := range rd.Entries {
					rdyEntries = append(rdyEntries, e)
					switch e.Type {
					case raftpb.EntryNormal:
					case raftpb.EntryConfChange:
						var cc raftpb.ConfChange
						cc.Unmarshal(e.Data)
						n2.ApplyConfChange(cc)
						applied = true
					}
				}

				log.Debug("n2.Ready", "msgs", rd.Messages, "entries", rd.Entries)
				for _, msg := range rd.Messages {
					nt.send(msg)
				}

				n2.Advance()
				if applied {
					nodes, weights := r2.nodes(false)
					t.Logf("n2.Ready: conf-changed: nodes: %v weights: %v", nodes, weights)
					applyConfChan2 <- struct{}{}
				}
			}
		}
	}()

	myAssert.Equal(1, len(r.prs))

	log.Debug("to AddNode 1", "r.prs", r.prs, "r", r)

	cc1 := raftpb.ConfChange{Type: raftpb.ConfChangeAddNode, NodeID: 1, Weight: 1}
	// ccdata1, _ := cc1.Marshal()
	n.ProposeConfChange(context.TODO(), cc1)
	<-applyConfChan

	log.Debug("to AddNode 1 again", "r.prs", r.prs, "r", r)

	cc3 := raftpb.ConfChange{Type: raftpb.ConfChangeAddNode, NodeID: 1, Weight: 3}
	// try add the same node again
	n.ProposeConfChange(context.TODO(), cc3)
	<-applyConfChan

	// try add the same node again
	n.ProposeConfChange(context.TODO(), cc3)
	<-applyConfChan

	myAssert.Equal(1, len(r.prs))

	time.Sleep(100 * time.Millisecond)

	log.Debug("after AddNode 1", "r.prs", r.prs, "r", r)

	myAssert.Equal(r.id, r.lead)

	log.Debug("to AddNode 2", "r.prs", r.prs, "r", r)

	// the new node join should be ok
	cc2 := raftpb.ConfChange{Type: raftpb.ConfChangeAddNode, NodeID: 2, Weight: 1}
	// ccdata2, _ := cc2.Marshal()
	n.ProposeConfChange(context.TODO(), cc2)
	<-applyConfChan

	time.Sleep(100 * time.Millisecond)

	log.Debug("to Run Node 2")
	go n2.run(r2)

	time.Sleep(100 * time.Millisecond)

	<-applyConfChan2

	close(stop)
	<-done
	<-done2

	t.Logf("r.prs: %v", r.prs)
	log.Debug("after done", "r.prs", r.prs, "r2.prs", r2.prs)
	myAssert.Equal(2, len(r.prs))

	myAssert.Equal(2, len(r2.prs))
	myAssert.Equal(StateLeader, r.state)
	myAssert.Equal(StateFollower, r2.state)

	log.Debug("TestConfChangeAddNode: done")
}

func TestConfChangeAddNodeInvalidWeight(t *testing.T) {
	setupTest(t)
	defer teardownTest(t)

	log.Debug("TestConfChangeAddNode: start")

	myAssert := assert.New(t)

	confFunc := func(c *Config) {
		c.PreVote = true
		c.peers = make(map[uint64]uint32)
		c.peers[c.ID] = uint32(1)
	}

	nt := newNetworkWithConfig(confFunc, nil, nil)
	log.Debug("after newNetworkWithConfig", "nt", nt)
	t.Logf("nt: %v", nt)

	n := newNode()
	s := nt.storage[1]
	//r := newTestRaft(1, []uint64{1}, 10, 1, s)
	r := nt.peers[1].(*raft)
	go n.run(r)

	n2 := newNode()
	s2 := nt.storage[2]
	//r2 := newTestRaft(2, []uint64{1}, 10, 1, s)
	r2 := nt.peers[2].(*raft)

	n.Campaign(context.TODO())
	rdyEntries := make([]raftpb.Entry, 0)

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	ticker2 := time.NewTicker(time.Millisecond * 10)
	defer ticker2.Stop()

	done := make(chan struct{})
	done2 := make(chan struct{})
	stop := make(chan struct{})
	applyConfChan := make(chan struct{})
	applyConfChan2 := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-stop: // stop
				return
			case <-ticker.C: // tick
				log.Debug("n.Tick: start")
				n.Tick()
				log.Debug("n.Tick: done")
			case rd := <-n.Ready(): // ready
				s.Append(rd.Entries)
				applied := false
				for _, e := range rd.Entries {
					rdyEntries = append(rdyEntries, e)
					switch e.Type {
					case raftpb.EntryNormal:
					case raftpb.EntryConfChange:
						var cc raftpb.ConfChange
						cc.Unmarshal(e.Data)
						n.ApplyConfChange(cc)
						applied = true
					}
				}
				log.Debug("n.Ready", "msgs", rd.Messages)
				for _, msg := range rd.Messages {
					nt.send(msg)
				}

				n.Advance()
				if applied {
					nodes, weights := r2.nodes(false)
					t.Logf("n.Ready: conf-changed: nodes: %v weights: %v", nodes, weights)
					applyConfChan <- struct{}{}
				}
			}
		}
	}()

	go func() {
		defer close(done2)
		for {
			select {
			case <-stop: // stop
				return
			case <-ticker2.C: // tick
				log.Debug("n2.Tick: start")
				n2.Tick()
				log.Debug("n2.Tick: done")
			case rd := <-n2.Ready(): // ready
				s2.Append(rd.Entries)
				applied := false
				for _, e := range rd.Entries {
					rdyEntries = append(rdyEntries, e)
					switch e.Type {
					case raftpb.EntryNormal:
					case raftpb.EntryConfChange:
						var cc raftpb.ConfChange
						cc.Unmarshal(e.Data)
						n2.ApplyConfChange(cc)
						applied = true
					}
				}

				log.Debug("n2.Ready", "msgs", rd.Messages, "entries", rd.Entries)
				for _, msg := range rd.Messages {
					nt.send(msg)
				}

				n2.Advance()
				if applied {
					nodes, weights := r2.nodes(false)
					t.Logf("n2.Ready: conf-changed: nodes: %v weights: %v", nodes, weights)
					applyConfChan2 <- struct{}{}
				}
			}
		}
	}()

	myAssert.Equal(1, len(r.prs))

	log.Debug("to AddNode 1", "r.prs", r.prs, "r", r)

	cc1 := raftpb.ConfChange{Type: raftpb.ConfChangeAddNode, NodeID: 1, Weight: 1}
	// ccdata1, _ := cc1.Marshal()
	n.ProposeConfChange(context.TODO(), cc1)
	<-applyConfChan

	log.Debug("to AddNode 1 again", "r.prs", r.prs, "r", r)

	cc3 := raftpb.ConfChange{Type: raftpb.ConfChangeAddNode, NodeID: 1, Weight: 3}
	// try add the same node again
	n.ProposeConfChange(context.TODO(), cc3)
	<-applyConfChan

	// try add the same node again
	n.ProposeConfChange(context.TODO(), cc3)
	<-applyConfChan

	myAssert.Equal(1, len(r.prs))

	time.Sleep(100 * time.Millisecond)

	log.Debug("after AddNode 1", "r.prs", r.prs, "r", r)

	myAssert.Equal(r.id, r.lead)

	log.Debug("to AddNode 2", "r.prs", r.prs, "r", r)

	// the new node join should be ok
	cc2 := raftpb.ConfChange{Type: raftpb.ConfChangeAddNode, NodeID: 2, Weight: 3}
	// ccdata2, _ := cc2.Marshal()
	err := n.ProposeConfChange(context.TODO(), cc2)
	t.Logf("after AddNode2: e: %v", err)
	/*
		if err == nil {
			<-applyConfChan
		}
	*/

	time.Sleep(100 * time.Millisecond)

	log.Debug("to Run Node 2")
	go n2.run(r2)

	time.Sleep(100 * time.Millisecond)

	close(stop)
	<-done
	<-done2

	t.Logf("r.prs: %v", r.prs)
	log.Debug("after done", "r.prs", r.prs, "r2.prs", r2.prs)
	myAssert.Equal(1, len(r.prs))

	myAssert.Equal(1, len(r2.prs))
	myAssert.Equal(StateLeader, r.state)
	myAssert.Equal(StateLeader, r2.state)

	log.Debug("TestConfChangeAddNode: done")
}
