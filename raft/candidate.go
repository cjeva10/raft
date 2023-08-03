package raft

import (
	"fmt"
	"sync"
)

func (n *Node) callElection() {
	votesNeeded := len(n.PeerList)/2 + 1
	votesReceived := 1
	votesFinished := 0

	n.mu.Lock()

	n.CurrentTerm++
	n.VotedFor = n.Id
	n.State = CANDIDATE
	n.LeaderId = 0

	// start new election timer
	go n.resetTimer()

	fmt.Printf("%v: Calling an election: Term %v\n", n.Id, n.CurrentTerm)
	n.mu.Unlock()

	// use a conditional to check vote totals later
	mu := new(sync.Mutex)
	cond := sync.NewCond(mu)

	// request votes from all peers
	for _, peer := range n.PeerList {
		thisPeer := peer
		go func() {
			n.mu.Lock()
			if n.State != CANDIDATE {
				n.mu.Unlock()
				return
			}
			n.mu.Unlock()

			fmt.Printf("%v: Requesting vote from %v\n", n.Id, thisPeer)
			vote := n.callRequestVote(thisPeer)

			mu.Lock()
			defer mu.Unlock()
			if vote {
				votesReceived++
				fmt.Printf("%v: Vote from %v, votesReceived: %v\n", n.Id, thisPeer, votesReceived)
			}
			votesFinished++
			cond.Broadcast()
		}()
	}

	mu.Lock()
	defer mu.Unlock()
	for votesReceived < votesNeeded && votesReceived <= len(n.PeerList) {
		cond.Wait()
	}
	if votesReceived >= votesNeeded {
		if n.State == CANDIDATE {
			go n.leader()
		}
	}
}

// if you're a candidate, you can call RequestVote to other nodes
func (n *Node) callRequestVote(peer int) bool {
	n.mu.Lock()
	argsTerm := n.CurrentTerm
	n.mu.Unlock()

	args := RequestVoteArgs{
		Term:        argsTerm,
		CandidateId: n.Id,
	}

	reply := RequestVoteReply{}

	ok := n.call(peer, "Node.RequestVote", &args, &reply)
	if !ok {
		return false
	}

	// another node has a higher term, therefore go back to follower
	if reply.Term > argsTerm {
		n.mu.Lock()
		n.becomeFollower(reply.Term)
		n.mu.Unlock()
		return false
	}

	// stale reply
	if reply.Term < argsTerm {
		return false
	}

	// we received a vote
	if reply.VoteGranted {
		return true
	}

	return false
}
