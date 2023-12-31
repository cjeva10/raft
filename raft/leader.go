package raft

import (
	"errors"
	"fmt"
	"os"
	// "strconv"
	"time"
)

// become the leader
func (n *Node) leader() {
	fmt.Printf("%v: Becoming the leader\n", n.Id)

	n.mu.Lock()
	n.State = LEADER
	n.Pulses++

	// initialize NextIndex
	for _, peer := range n.PeerList {
		n.NextIndex[peer] = len(n.Log)
		n.MatchIndex[peer] = 0
	}

	n.mu.Unlock()

	// start a thread to update the commit index
	if !n.Killed {
		go n.commitCheck()
	}

	for !n.Killed {
		// the leader's pulse is quicker than the follower's
		time.Sleep(PULSETIME / 2 * time.Millisecond)

		fmt.Printf("%v: sending heartbeats\n", n.Id)
		go n.addEntries()

		fmt.Printf("%v: Current Log: %v\n", n.Id, n.Log)
	}
}

func (n *Node) commitCheck() {
	n.mu.Lock()
	defer n.mu.Unlock()

	N := n.CommitIndex + 1
	// we need to broadcast if any of these values are changed and we can still possibly be the leader
	for N <= n.CommitIndex || !n.majorityMatchIndex(N) || len(n.Log) <= N || n.Log[N].Term != n.CurrentTerm {
		n.CommitCond.Wait()
	}
	n.CommitIndex = N
	n.checkLastApplied()
	fmt.Printf("%v: New CommitIndex = %v\n", n.Id, N)
	n.CommitCond.Broadcast()

	// calling this function recursively ensures that CommitIndex increases monotonically,
	// because there can't be two threads on this function at the same time
	// i.e. this thread calls commitCheck, which waits for the lock
	// this thread then returns and releases the lock, so the new thread can't do anything
	// until this one returns
	if !n.Killed && n.State == LEADER {
		go n.commitCheck()
	}
}

// we're already holding the lock when this loop is called
// check whether a majority of matchIndex[i] >= N
func (n *Node) majorityMatchIndex(N int) bool {
	majority := len(n.PeerList)/2 + 1
	count := 1
	for _, match := range n.MatchIndex {
		if match >= N {
			count++
		}
	}

	if count >= majority {
		return true
	}
	return false
}

type ClientRequestArgs struct {
	Command string
}
type ClientRequestReply struct {
	Success bool
}

// RPC for receiving client requests
// for now assume only the leader can receive client requests
// fix that issue later
func (n *Node) ClientRequest(args *ClientRequestArgs, reply *ClientRequestReply) error {
	n.mu.Lock()
	// only the leader can receive client requests
	if n.State != LEADER {
		n.mu.Unlock()
		reply.Success = false
		return errors.New("I'm not the leader")
	}

	entry := LogEntry{
		Command: args.Command,
		Term:    n.CurrentTerm,
	}
	n.Log = append(n.Log, entry)
	idx := len(n.Log) - 1
	n.CommitCond.Broadcast()
	n.mu.Unlock()

	n.mu.Lock()
	// wait until commit index is greater equal than index of this log
	for n.CommitIndex < idx {
		n.CommitCond.Wait()
	}
	n.mu.Unlock()

	// then reply to the client
	reply.Success = true
	return nil
}

type LeaderAppendReply struct {
	Aer  AppendEntriesReply
	Peer int
}

// routine for a leader to commit a log entry
func (n *Node) addEntries() {
	// send AppendEntries to each of the peers
	allArgs := []AppendEntriesArgs{}
	if n.Killed {
		return
	}

	n.mu.Lock()
	// make sure that we are still the leader
	if n.State != LEADER {
		return
	}

	lastIdx := len(n.Log) - 1
	thisTerm := n.CurrentTerm
	for _, peer := range n.PeerList {
		var entries []LogEntry
		prevLogIndex := lastIdx
		prevLogTerm := n.Log[prevLogIndex].Term

		// if last log index >= nextIndex for a follower...
		if lastIdx >= n.NextIndex[peer] {
			// call AppendEntries with log entries starting at next index
			entries = n.Log[n.NextIndex[peer]:]
			prevLogIndex = n.NextIndex[peer] - 1
			prevLogTerm = n.Log[prevLogIndex].Term
		}

		args := AppendEntriesArgs{
			Term:         thisTerm,
			LeaderId:     n.Id,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      entries,
			LeaderCommit: n.CommitIndex,
		}

		allArgs = append(allArgs, args)
	}
	n.mu.Unlock()

	larch := make(chan LeaderAppendReply, len(n.PeerList))
	for i, peer := range n.PeerList {
		fmt.Printf("%v: Calling AppendEntries to %v\n", n.Id, peer)
		n.mu.Lock()
		if n.State == LEADER {
			n.mu.Unlock()
			go n.callAppendEntries(peer, allArgs[i], larch)
		} else {
			n.mu.Unlock()
			return
		}
	}

	for reply := range larch {
		n.mu.Lock()
		fmt.Printf("%v Term %v: AppendEntriesReply: Id %v, Term %v, Success %v\n", n.Id, n.CurrentTerm, reply.Peer, reply.Aer.Term, reply.Aer.Success)

		// if reply term > current term, immediately become a folower and kill function
		if reply.Aer.Term > thisTerm {
			n.becomeFollower(reply.Aer.Term)
			n.mu.Unlock()
			return
		}

		// successful reply, update NextIndex and MatchIndex for this peer
		if reply.Aer.Success {
			n.NextIndex[reply.Peer] = lastIdx + 1
			n.MatchIndex[reply.Peer] = lastIdx
			n.CommitCond.Broadcast()
			if len(n.Log)-1 > lastIdx {
				fmt.Fprintf(os.Stderr, "Log length increased between locks! new %v vs old %v\n", len(n.Log)-1, lastIdx)
			}

			n.mu.Unlock()
		} else {
			// only failure case is inconsistent log -> decrement NextIndex and retry
			n.NextIndex[reply.Peer]--
			entries := n.Log[n.NextIndex[reply.Peer]:]
			prevLogIndex := n.NextIndex[reply.Peer] - 1
			prevLogTerm := n.Log[prevLogIndex].Term

			args := AppendEntriesArgs{
				Term:         n.CurrentTerm,
				LeaderId:     n.Id,
				PrevLogIndex: prevLogIndex,
				PrevLogTerm:  prevLogTerm,
				Entries:      entries,
				LeaderCommit: n.CommitIndex,
			}

			// retry
			n.mu.Unlock()
			fmt.Printf("%v: Bad reply, resending to %v\n", n.Id, reply.Peer)
			go n.callAppendEntries(reply.Peer, args, larch)
		}
	}
}

// if you're the leader you can send AppendEntries to other nodes
func (n *Node) callAppendEntries(peer int, args AppendEntriesArgs, larch chan LeaderAppendReply) {
	reply := AppendEntriesReply{}

	ok := n.call(peer, "Node.AppendEntries", &args, &reply)
	if ok {
		larch <- LeaderAppendReply{
			Aer:  reply,
			Peer: peer,
		}
	}
}
