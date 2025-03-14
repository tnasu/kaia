// Modifications Copyright 2024 The Kaia Authors
// Modifications Copyright 2018 The klaytn Authors
// Copyright 2014 The go-ethereum Authors
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
//
// This file is derived from p2p/peer_test.go (2018/06/04).
// Modified and improved for the klaytn development.
// Modified and improved for the Kaia development.

package p2p

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/kaiachain/kaia/common/math"
	"github.com/stretchr/testify/assert"
)

var discard = Protocol{
	Name:   "discard",
	Length: 1,
	Run: func(p *Peer, rw MsgReadWriter) error {
		for {
			msg, err := rw.ReadMsg()
			if err != nil {
				return err
			}
			fmt.Printf("discarding %d\n", msg.Code)
			if err = msg.Discard(); err != nil {
				return err
			}
		}
	},
}

func testPeer(protos []Protocol) (func(), *conn, *Peer, <-chan error) {
	var (
		fd1, fd2   = net.Pipe()
		id1, id2   = randomID(), randomID()
		pubkey1, _ = id1.Pubkey()
		t1         = newTestTransport(id2, fd1, nil, true)
		t2         = newTestTransport(id1, fd2, pubkey1, true)
	)
	c1 := &conn{fd: fd1, transport: t1}
	c2 := &conn{fd: fd2, transport: t2}
	for _, p := range protos {
		c1.caps = append(c1.caps, p.cap())
		c2.caps = append(c2.caps, p.cap())
	}

	peer, _ := newPeer([]*conn{c1}, protos, defaultRWTimerConfig)
	errc := make(chan error, 1)
	go func() {
		_, err := peer.run()
		errc <- err
	}()

	closer := func() { c2.close(errors.New("close func called")) }
	return closer, c2, peer, errc
}

func testPeerWithRWs(protos []Protocol, channelSize int) (func(), []*conn, *Peer, <-chan error) {
	serverSideConn := make([]*conn, 0, channelSize)
	peerSideConn := make([]*conn, 0, channelSize)

	for i := 0; i < channelSize; i++ {
		var (
			fd1, fd2   = net.Pipe()
			id1, id2   = randomID(), randomID()
			pubkey1, _ = id1.Pubkey()
			t1         = newTestTransport(id2, fd1, nil, true)
			t2         = newTestTransport(id1, fd2, pubkey1, true)
		)
		c1 := &conn{fd: fd1, transport: t1}
		c2 := &conn{fd: fd2, transport: t2}
		for _, p := range protos {
			c1.caps = append(c1.caps, p.cap())
			c2.caps = append(c2.caps, p.cap())
		}
		serverSideConn = append(serverSideConn, c1)
		peerSideConn = append(peerSideConn, c2)
	}

	peer, _ := newPeer(serverSideConn, protos, defaultRWTimerConfig)
	errc := make(chan error, 1)
	go func() {
		_, err := peer.runWithRWs()
		errc <- err
	}()

	closer := func() {
		for _, conn := range peerSideConn {
			conn.close(errors.New("close func called"))
		}
	}
	return closer, peerSideConn, peer, errc
}

func TestPeerProtoReadMsg(t *testing.T) {
	proto := Protocol{
		Name:   "a",
		Length: 5,
		Run: func(peer *Peer, rw MsgReadWriter) error {
			if err := ExpectMsg(rw, 2, []uint{1}); err != nil {
				t.Error(err)
			}
			if err := ExpectMsg(rw, 3, []uint{2}); err != nil {
				t.Error(err)
			}
			if err := ExpectMsg(rw, 4, []uint{3}); err != nil {
				t.Error(err)
			}
			return nil
		},
	}

	closer, rw, _, errc := testPeer([]Protocol{proto})
	defer closer()

	Send(rw, baseProtocolLength+2, []uint{1})
	Send(rw, baseProtocolLength+3, []uint{2})
	Send(rw, baseProtocolLength+4, []uint{3})

	select {
	case err := <-errc:
		if err != errProtocolReturned {
			t.Errorf("peer returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("receive timeout")
	}
}

func TestPeerProtoEncodeMsg(t *testing.T) {
	proto := Protocol{
		Name:   "a",
		Length: 2,
		Run: func(peer *Peer, rw MsgReadWriter) error {
			if err := SendItems(rw, 2); err == nil {
				t.Error("expected error for out-of-range msg code, got nil")
			}
			if err := SendItems(rw, 1, "foo", "bar"); err != nil {
				t.Errorf("write error: %v", err)
			}
			return nil
		},
	}
	closer, rw, _, _ := testPeer([]Protocol{proto})
	defer closer()

	if err := ExpectMsg(rw, 17, []string{"foo", "bar"}); err != nil {
		t.Error(err)
	}
}

func TestPeerPing(t *testing.T) {
	closer, rw, _, _ := testPeer(nil)
	defer closer()
	if err := SendItems(rw, pingMsg); err != nil {
		t.Fatal(err)
	}
	if err := ExpectMsg(rw, pongMsg, nil); err != nil {
		t.Error(err)
	}
}

func TestPeerDisconnect(t *testing.T) {
	for _, tc := range []struct {
		dc    DiscReason
		error string
	}{
		{DiscQuitting, "client quitting"},
		{DiscSubprotocolError, "subprotocol error"},
		{17, "unknown disconnect reason 17"},
		{18, "unknown disconnect reason 18"},
		{math.MaxUint8, "unknown disconnect reason 255"},
	} {
		closer, rw, _, disc := testPeer(nil)
		if err := SendItems(rw, discMsg, tc.dc); err != nil {
			t.Fatal(err)
		}
		select {
		case reason := <-disc:
			assert.Equal(t, tc.dc, reason)
			assert.Equal(t, tc.error, reason.Error())
		case <-time.After(500 * time.Millisecond):
			t.Error("peer did not return")
		}
		closer()
	}
}

// This test is supposed to verify that Peer can reliably handle
// multiple causes of disconnection occurring at the same time.
func TestPeerDisconnectRace(t *testing.T) {
	maybe := func() bool { return rand.Intn(2) == 1 }

	for i := 0; i < 1000; i++ {
		protoclose := make(chan error)
		protodisc := make(chan DiscReason)
		closer, rw, p, disc := testPeer([]Protocol{
			{
				Name:   "closereq",
				Run:    func(p *Peer, rw MsgReadWriter) error { return <-protoclose },
				Length: 1,
			},
			{
				Name:   "disconnect",
				Run:    func(p *Peer, rw MsgReadWriter) error { p.Disconnect(<-protodisc); return nil },
				Length: 1,
			},
		})

		// Simulate incoming messages.
		go SendItems(rw, baseProtocolLength+1)
		go SendItems(rw, baseProtocolLength+2)
		// Close the network connection.
		go closer()
		// Make protocol "closereq" return.
		protoclose <- errors.New("protocol closed")
		// Make protocol "disconnect" call peer.Disconnect
		protodisc <- DiscAlreadyConnected
		// In some cases, simulate something else calling peer.Disconnect.
		if maybe() {
			go p.Disconnect(DiscInvalidIdentity)
		}
		// In some cases, simulate remote requesting a disconnect.
		if maybe() {
			go SendItems(rw, discMsg, DiscQuitting)
		}

		select {
		case <-disc:
		case <-time.After(2 * time.Second):
			// Peer.run should return quickly. If it doesn't the Peer
			// goroutines are probably deadlocked. Call panic in order to
			// show the stacks.
			panic("Peer.run took to long to return.")
		}
	}
}

func TestMultiChannelPeerProtoReadMsg(t *testing.T) {
	proto := Protocol{
		Name:   "a",
		Length: 5,
		RunWithRWs: func(peer *Peer, rws []MsgReadWriter) error {
			for _, rw := range rws {
				if err := ExpectMsg(rw, 2, []uint{1}); err != nil {
					t.Error(err)
				}
				if err := ExpectMsg(rw, 3, []uint{2}); err != nil {
					t.Error(err)
				}
				if err := ExpectMsg(rw, 4, []uint{3}); err != nil {
					t.Error(err)
				}
			}
			return nil
		},
	}

	closer, rws, _, errc := testPeerWithRWs([]Protocol{proto}, 2)
	defer closer()

	for _, rw := range rws {
		Send(rw, baseProtocolLength+2, []uint{1})
		Send(rw, baseProtocolLength+3, []uint{2})
		Send(rw, baseProtocolLength+4, []uint{3})
	}

	select {
	case err := <-errc:
		if err != errProtocolReturned {
			t.Errorf("peer returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("receive timeout")
	}
}

func TestMultiChannelPeerProtoEncodeMsg(t *testing.T) {
	proto := Protocol{
		Name:   "a",
		Length: 2,
		RunWithRWs: func(peer *Peer, rws []MsgReadWriter) error {
			for _, rw := range rws {
				if err := SendItems(rw, 2); err == nil {
					t.Error("expected error for out-of-range msg code, got nil")
				}
				if err := SendItems(rw, 1, "foo", "bar"); err != nil {
					t.Errorf("write error: %v", err)
				}
			}
			return nil
		},
	}
	closer, rws, _, _ := testPeerWithRWs([]Protocol{proto}, 2)
	defer closer()

	for _, rw := range rws {
		if err := ExpectMsg(rw, 17, []string{"foo", "bar"}); err != nil {
			t.Error(err)
		}
	}
}

func TestMultiChannelPeerPing(t *testing.T) {
	closer, rws, _, _ := testPeerWithRWs(nil, 2)
	defer closer()

	for _, rw := range rws {
		if err := SendItems(rw, pingMsg); err != nil {
			t.Fatal(err)
		}
		if err := ExpectMsg(rw, pongMsg, nil); err != nil {
			t.Error(err)
		}
	}
}

func TestMultiChannelPeerDisconnect(t *testing.T) {
	channelSize := 2
	for i := 0; i < channelSize; i++ {
		closer, rws, _, disc := testPeerWithRWs(nil, channelSize)
		defer closer()

		if err := SendItems(rws[i], discMsg, DiscQuitting); err != nil {
			t.Fatal(err)
		}

		select {
		case reason := <-disc:
			if reason != DiscQuitting {
				t.Errorf("run returned wrong reason: got %v, want %v", reason, DiscQuitting)
			}
		case <-time.After(500 * time.Millisecond):
			t.Error("peer did not return")
		}
	}
}

// This test is supposed to verify that Peer can reliably handle
// multiple causes of disconnection occurring at the same time.
func TestMultiChannelPeerDisconnectRace(t *testing.T) {
	maybe := func() bool { return rand.Intn(2) == 1 }
	channelSize := 2

	for i := 0; i < 1000; i++ {
		protoclose := make(chan error)
		protodisc := make(chan DiscReason)
		closer, rws, p, disc := testPeerWithRWs([]Protocol{
			{
				Name:       "closereq",
				RunWithRWs: func(p *Peer, rw []MsgReadWriter) error { return <-protoclose },
				Length:     1,
			},
			{
				Name:       "disconnect",
				RunWithRWs: func(p *Peer, rw []MsgReadWriter) error { p.Disconnect(<-protodisc); return nil },
				Length:     1,
			},
		}, channelSize)

		// Simulate incoming messages.
		for _, rw := range rws {
			go SendItems(rw, baseProtocolLength+1)
			go SendItems(rw, baseProtocolLength+2)
		}
		// Close the network connection.
		go closer()
		// Make protocol "closereq" return.
		protoclose <- errors.New("protocol closed")
		// Make protocol "disconnect" call peer.Disconnect
		protodisc <- DiscAlreadyConnected
		// In some cases, simulate something else calling peer.Disconnect.
		if maybe() {
			go p.Disconnect(DiscInvalidIdentity)
		}
		// In some cases, simulate remote requesting a disconnect.
		if maybe() {
			go SendItems(rws[rand.Intn(channelSize)], discMsg, DiscQuitting)
		}

		select {
		case <-disc:
		case <-time.After(2 * time.Second):
			// Peer.run should return quickly. If it doesn't the Peer
			// goroutines are probably deadlocked. Call panic in order to
			// show the stacks.
			panic("Peer.run took to long to return.")
		}
	}
}

func TestNewPeer(t *testing.T) {
	name := "nodename"
	caps := []Cap{{"foo", 2}, {"bar", 3}}
	id := randomID()
	p := NewPeer(id, name, caps)
	if p.ID() != id {
		t.Errorf("ID mismatch: got %v, expected %v", p.ID(), id)
	}
	if p.Name() != name {
		t.Errorf("Name mismatch: got %v, expected %v", p.Name(), name)
	}
	if !reflect.DeepEqual(p.Caps(), caps) {
		t.Errorf("Caps mismatch: got %v, expected %v", p.Caps(), caps)
	}

	p.Disconnect(DiscAlreadyConnected) // Should not hang
}

func TestMatchProtocols(t *testing.T) {
	tests := []struct {
		Remote []Cap
		Local  []Protocol
		Match  map[string]protoRW
	}{
		{
			// No remote capabilities
			Local: []Protocol{{Name: "a"}},
		},
		{
			// No local protocols
			Remote: []Cap{{Name: "a"}},
		},
		{
			// No mutual protocols
			Remote: []Cap{{Name: "a"}},
			Local:  []Protocol{{Name: "b"}},
		},
		{
			// Some matches, some differences
			Remote: []Cap{{Name: "local"}, {Name: "match1"}, {Name: "match2"}},
			Local:  []Protocol{{Name: "match1"}, {Name: "match2"}, {Name: "remote"}},
			Match: map[string]protoRW{
				"match1": {Protocol: Protocol{Name: "match1"}, tc: defaultRWTimerConfig},
				"match2": {Protocol: Protocol{Name: "match2"}, tc: defaultRWTimerConfig},
			},
		},
		{
			// Various alphabetical ordering
			Remote: []Cap{{Name: "aa"}, {Name: "ab"}, {Name: "bb"}, {Name: "ba"}},
			Local:  []Protocol{{Name: "ba"}, {Name: "bb"}, {Name: "ab"}, {Name: "aa"}},
			Match: map[string]protoRW{
				"aa": {Protocol: Protocol{Name: "aa"}, tc: defaultRWTimerConfig},
				"ab": {Protocol: Protocol{Name: "ab"}, tc: defaultRWTimerConfig},
				"ba": {Protocol: Protocol{Name: "ba"}, tc: defaultRWTimerConfig},
				"bb": {Protocol: Protocol{Name: "bb"}, tc: defaultRWTimerConfig},
			},
		},
		{
			// No mutual versions
			Remote: []Cap{{Version: 1}},
			Local:  []Protocol{{Version: 2}},
		},
		{
			// Multiple versions, single common
			Remote: []Cap{{Version: 1}, {Version: 2}},
			Local:  []Protocol{{Version: 2}, {Version: 3}},
			Match:  map[string]protoRW{"": {Protocol: Protocol{Version: 2}, tc: defaultRWTimerConfig}},
		},
		{
			// Multiple versions, multiple common
			Remote: []Cap{{Version: 1}, {Version: 2}, {Version: 3}, {Version: 4}},
			Local:  []Protocol{{Version: 2}, {Version: 3}},
			Match:  map[string]protoRW{"": {Protocol: Protocol{Version: 3}, tc: defaultRWTimerConfig}},
		},
		{
			// Various version orderings
			Remote: []Cap{{Version: 4}, {Version: 1}, {Version: 3}, {Version: 2}},
			Local:  []Protocol{{Version: 2}, {Version: 3}, {Version: 1}},
			Match:  map[string]protoRW{"": {Protocol: Protocol{Version: 3}, tc: defaultRWTimerConfig}},
		},
		{
			// Versions overriding sub-protocol lengths
			Remote: []Cap{{Version: 1}, {Version: 2}, {Version: 3}, {Name: "a"}},
			Local:  []Protocol{{Version: 1, Length: 1}, {Version: 2, Length: 2}, {Version: 3, Length: 3}, {Name: "a"}},
			Match: map[string]protoRW{
				"":  {Protocol: Protocol{Version: 3}, tc: defaultRWTimerConfig},
				"a": {Protocol: Protocol{Name: "a"}, offset: 3, tc: defaultRWTimerConfig},
			},
		},
	}

	for i, tt := range tests {
		result := matchProtocols(tt.Local, tt.Remote, nil, defaultRWTimerConfig)
		if len(result) != len(tt.Match) {
			t.Errorf("test %d: negotiation mismatch: have %v, want %v", i, len(result), len(tt.Match))
			continue
		}
		// Make sure all negotiated protocols are needed and correct
		for name, proto := range result {
			match, ok := tt.Match[name]
			if !ok {
				t.Errorf("test %d, protobuf '%s': negotiated but shouldn't have", i, name)
				continue
			}
			if proto[ConnDefault].Name != match.Name {
				t.Errorf("test %d, protobuf '%s': name mismatch: have %v, want %v", i, name, proto[ConnDefault].Name, match.Name)
			}
			if proto[ConnDefault].Version != match.Version {
				t.Errorf("test %d, protobuf '%s': version mismatch: have %v, want %v", i, name, proto[ConnDefault].Version, match.Version)
			}
			if proto[ConnDefault].offset-baseProtocolLength != match.offset {
				t.Errorf("test %d, protobuf '%s': offset mismatch: have %v, want %v", i, name, proto[ConnDefault].offset-baseProtocolLength, match.offset)
			}
		}
		// Make sure no protocols missed negotiation
		for name := range tt.Match {
			if _, ok := result[name]; !ok {
				t.Errorf("test %d, protobuf '%s': not negotiated, should have", i, name)
				continue
			}
		}
	}
}
