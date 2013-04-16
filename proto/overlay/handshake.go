// Iris - Distributed Messaging Framework
// Copyright 2013 Peter Szilagyi. All rights reserved.
//
// Iris is dual licensed: you can redistribute it and/or modify it under the
// terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// The framework is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// Alternatively, the Iris framework may be used in accordance with the terms
// and conditions contained in a signed written agreement between you and the
// author(s).
//
// Author: peterke@gmail.com (Peter Szilagyi)
package overlay

import (
	"config"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"net"
	"proto/bootstrap"
	"proto/session"
	"sort"
	"time"
)

// The initialization packet when the connection is set up.
type initPacket struct {
	Id    *big.Int
	Addrs []string
}

// Starts up the overlay networking on a specified interface and fans in all the
// inbound connections into the overlay-global channels, tagging them with the
// source interface.
func (o *overlay) acceptor(ip net.IP) {
	// Listen for incomming session on the given interface and random port.
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(ip.String(), "0"))
	if err != nil {
		panic(fmt.Sprintf("failed to resolve interface (%v): %v.", ip, err))
	}
	sesSink, quit, err := session.Listen(addr, o.lkey, o.rkeys)
	if err != nil {
		panic(fmt.Sprintf("failed to start session listener: %v.", err))
	}
	defer close(quit)

	// Save the new listener address into the local address list
	o.mutex.Lock()
	o.addrs = append(o.addrs, addr.String())
	sort.Sort(sort.StringSlice(o.addrs))
	o.mutex.Unlock()

	// Start the bootstrapper on the specified interface
	bootSink, quit, err := bootstrap.Boot(ip, []byte(o.overId), addr.Port)
	if err != nil {
		panic(fmt.Sprintf("failed to start bootstrapper: %v.", err))
	}
	defer close(quit)

	// Loop indefinitely, faning in the sessions and discovered peers
	for {
		select {
		case <-o.quit:
			return
		case boot := <-bootSink:
			o.bootSink <- boot
		case ses := <-sesSink:
			o.sesSink <- ses
		}
	}
}

// Filters out inbound connections, and executes the handshake for those that
// are deemed useful.
func (o *overlay) shaker() {
	for {
		select {
		case <-o.quit:
			return
		case boot := <-o.bootSink:
			// Connect'em all (for now) !!!
			go func() {
				if ses, err := session.Dial(boot.IP.String(), boot.Port, o.overId, o.lkey, o.rkeys[o.overId]); err != nil {
					log.Printf("failed to dial remote pastry peer: %v.", err)
				} else {
					o.sesSink <- ses
				}
			}()
		case ses := <-o.sesSink:
			// Wait for peer init packet for real address
			go o.shake(ses)
		}
	}
}

// Pastry handshake to sort out the real hosts and ports.
func (o *overlay) shake(ses *session.Session) {
	// Create a new peer structure to hold all the needed session data
	p := new(peer)

	p.laddr = ses.Raw().LocalAddr().String()
	p.raddr = ses.Raw().RemoteAddr().String()
	p.dec = gob.NewDecoder(&p.inBuf)
	p.enc = gob.NewEncoder(&p.outBuf)

	p.quit = make(chan struct{})
	p.in = make(chan *session.Message)
	p.out = ses.Communicate(p.in, p.quit)

	// Send an init packet to the remote peer
	pkt := new(initPacket)
	pkt.Id = new(big.Int).Set(o.nodeId)

	o.mutex.Lock()
	pkt.Addrs = make([]string, len(o.addrs))
	copy(pkt.Addrs, o.addrs)
	o.mutex.Unlock()

	if err := p.enc.Encode(pkt); err != nil {
		log.Printf("failed to encode init packet: %v.", err)
		return
	}
	msg := new(session.Message)
	msg.Head.Meta = make([]byte, p.outBuf.Len())
	copy(msg.Head.Meta, p.outBuf.Bytes())
	p.out <- msg

	// Wait for an incoming init packet
	timeout := time.Tick(time.Duration(config.PastryInitTimeout) * time.Millisecond)
	select {
	case <-timeout:
		log.Printf("session initialization timed out: %vms.", config.PastryInitTimeout)
		return
	case msg, ok := <-p.in:
		if !ok {
			log.Printf("remote closed connection before init packet.")
			return
		}
		p.inBuf.Write(msg.Head.Meta)
		if err := p.dec.Decode(pkt); err != nil {
			log.Printf("failed to decode remote init packet: %v.", err)
			return
		}
		p.self = pkt.Id
		p.addrs = pkt.Addrs

		// Everything ok, accept connection
		o.integrate(p)
	}
}
