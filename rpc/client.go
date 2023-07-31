package rpc

import (
	"fmt"
	"net/rpc"
	"strconv"
	"sync"
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

func (n *Node) callElection() {
	votesNeeded := len(n.PeerList)/2 + 1
	votesReceived := 1
	votesFinished := 0

	n.mu.Lock()

	n.CurrentTerm++
	n.VotedFor = n.Id
	n.State = CANDIDATE

	// start new election timer
	n.Pulses++
	go n.pulseCheck()

	fmt.Printf("Calling an election: Term %v\n", n.CurrentTerm)

	n.mu.Unlock()

	cond := sync.NewCond(&n.mu)

	// request votes from all peers
	for _, peer := range n.PeerList {
		thisPeer := peer
		go func() {
            fmt.Printf("Requesting vote from %v\n", thisPeer)
			vote := n.callRequestVote(thisPeer)
			n.mu.Lock()
			defer n.mu.Unlock()
            if n.State == FOLLOWER {
                // if we converted back to a follower cancel the election
                return
            }

			if vote {
                fmt.Printf("Received vote from %v\n", thisPeer)
				votesReceived++
			}
			votesFinished++
			cond.Broadcast()
		}()
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	for votesReceived < votesNeeded && votesReceived <= len(n.PeerList) {
        fmt.Printf("votes received = %v\n", votesReceived)
		cond.Wait()
	}
	if votesReceived >= votesNeeded {
		go n.leader()
	}
}

// if you're a candidate, you can call RequestVote to other nodes
func (n *Node) callRequestVote(peer int) bool {
    n.mu.Lock()
	args := RequestVoteArgs{
		Term:        n.CurrentTerm,
		CandidateId: n.Id,
	}
    n.mu.Unlock()

	reply := RequestVoteReply{}

	ok := call(peer, "Node.RequestVote", &args, &reply)
	if !ok {
		return false
	}

    n.mu.Lock()
    defer n.mu.Unlock()

	// another node has a higher term, therefore go back to follower
	if reply.Term > n.CurrentTerm {
		n.State = FOLLOWER
		n.Pulses++
		n.VotedFor = 0
        go n.pulseCheck()
		return false
	}

	// we received a vote
	if reply.VoteGranted {
		fmt.Println("We got a vote!")
		return true
	}

	return false
}

// if you're the leader you can send AppendEntries to other nodes
func (n *Node) callAppendEntries(peer int) {
	n.mu.Lock()
    idx := len(n.Log) - 1 
	args := AppendEntriesArgs{
		Term:         n.CurrentTerm,
		LeaderId:     n.Id,
		PrevLogIndex: idx,
		PrevLogTerm:  n.LogTerms[idx],
		Entries:      []string{},
		LeaderCommit: 0,
	}
	n.mu.Unlock()

	reply := AppendEntriesReply{}

	ok := call(peer, "Node.AppendEntries", args, &reply)
	if ok {
		// todo
	} else {
		// todo
	}

}
