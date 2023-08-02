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

	// start new election timer
	n.Pulses++
	go n.pulseCheck()

	fmt.Printf("Calling an election: Term %v\n", n.CurrentTerm)

	n.mu.Unlock()

	// use a conditional to check vote totals later
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

	ok := n.call(peer, "Node.RequestVote", &args, &reply)
	if !ok {
		return false
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	// another node has a higher term, therefore go back to follower
	if reply.Term > n.CurrentTerm {
        n.becomeFollower(reply.Term)
		return false
	}

	// we received a vote
	if reply.VoteGranted {
		fmt.Println("We got a vote!")
		return true
	}

	return false
}

