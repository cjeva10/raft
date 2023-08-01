package raft

import (
	"math/rand"
	"net/rpc"
	"strconv"
	"time"
)

// helper function to call RPC methods on a peer
func call(peer int, rpcname string, args interface{}, reply interface{}) bool {
	peername := strconv.Itoa(peer)

	c, err := rpc.DialHTTP("tcp", "localhost:"+peername)
	if err != nil {
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	return false
}

// wait for a random amount of time (150-300ms) and then request votes
func (n *Node) pulseCheck() {
	// kill all existing pulses (by incrementing counter) and record current total
	n.mu.Lock()
	n.Pulses++
	currPulses := n.Pulses
	n.mu.Unlock()

	// wait
	delay := 2000 + rand.Intn(500)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// if we didn't receive a pulse, call an election
	n.mu.Lock()
	if currPulses == n.Pulses {
		go n.callElection() // do this on another thread to release the lock
	}
	n.mu.Unlock()
}
