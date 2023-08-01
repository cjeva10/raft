package raft

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// become the leader
func (n *Node) leader() {
	fmt.Println("Becoming the leader")
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
	go n.commitCheck()

    // simulate two client requests
	args := ClientRequestArgs{
		Command: "hello",
	}
	reply := ClientRequestReply{}
	go n.ClientRequest(&args, &reply)

	args2 := ClientRequestArgs{
		Command: "world",
	}
	reply2 := ClientRequestReply{}
	go n.ClientRequest(&args2, &reply2)

	for {
		time.Sleep(1500 * time.Millisecond)
        n.mu.Lock()
		fmt.Printf("Current Log: %v\n", n.Log)
        n.mu.Unlock()

		go n.addEntries()

        // simulate an additional request each heartbeat
        args3 := ClientRequestArgs{
            Command: "third",
        }
        go n.ClientRequest(&args3, &ClientRequestReply{})
        go n.ClientRequest(&args2, &ClientRequestReply{})
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
	n.CommitCond.Broadcast()

	// calling this function recursively ensures that CommitIndex increases monotonically,
	// because there can't be two threads on this function at the same time
	// i.e. this thread calls commitCheck, which waits for the lock
	// this thread then returns and releases the lock, so the new thread can't do anything
	// until this one returns
	go n.commitCheck()
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
	larch := make(chan LeaderAppendReply)
	for _, peer := range n.PeerList {
		n.mu.Lock()
		// if last log index >= nextIndex for a follower...
		var entries []LogEntry
		prevLogIndex := len(n.Log) - 1
		prevLogTerm := n.Log[prevLogIndex].Term
		if len(n.Log)-1 >= n.NextIndex[peer] {
			// call AppendEntries with log entries starting at next index
			entries = n.Log[n.NextIndex[peer]:]
			prevLogIndex = n.NextIndex[peer] - 1
			prevLogTerm = n.Log[prevLogIndex].Term
		}
		go n.callAppendEntries(peer, prevLogIndex, prevLogTerm, entries, larch)
		n.mu.Unlock()
	}

	for reply := range larch {
		n.mu.Lock() // lock for the term check
		// if reply term > current term, immediately become a folower and kill function
		if reply.Aer.Term > n.CurrentTerm {
			n.becomeFollower(reply.Aer.Term)
			n.mu.Unlock()
			return
		}
		n.mu.Unlock()

		// successful reply, update NextIndex and MatchIndex for this peer
		if reply.Aer.Success {
			n.mu.Lock()

			n.NextIndex[reply.Peer] = len(n.Log)
			n.MatchIndex[reply.Peer] = len(n.Log) - 1
			n.CommitCond.Broadcast()

			n.mu.Unlock()
			// only failure case is inconsistent log -> decrement NextIndex and retry
		} else {
			n.mu.Lock()
			n.NextIndex[reply.Peer]--
			entries := n.Log[n.NextIndex[reply.Peer]:]
			prevLogIndex := n.NextIndex[reply.Peer]
			prevLogTerm := n.Log[prevLogIndex].Term
			n.mu.Unlock()

			// retry
			go n.callAppendEntries(reply.Peer, prevLogIndex, prevLogTerm, entries, larch)
		}
	}
}

// if you're the leader you can send AppendEntries to other nodes
func (n *Node) callAppendEntries(peer int, prevLogIndex int, prevLogTerm int, entries []LogEntry, larch chan LeaderAppendReply) {
	n.mu.Lock()
	// fmt.Printf("calling append entries to: %v\n", peer)
	// fmt.Printf("Term: %v\n", n.CurrentTerm)
	// fmt.Printf("LeaderId: %v\n", n.Id)
	// fmt.Printf("PrevLogIndex: %v\n", prevLogIndex)
	// fmt.Printf("PrevLogTerm: %v\n", prevLogTerm)
	// fmt.Printf("Entries: %v\n", entries)
	// fmt.Printf("LeaderCommit: %v\n", n.CommitIndex)
	// fmt.Printf("\n")

	args := AppendEntriesArgs{
		Term:         n.CurrentTerm,
		LeaderId:     n.Id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: n.CommitIndex,
	}
	n.mu.Unlock()

	reply := AppendEntriesReply{}

	ok := call(peer, "Node.AppendEntries", args, &reply)
	if !ok {
		fmt.Fprintf(os.Stderr, "Failed to receive RPC request\n")
	} else {
		larch <- LeaderAppendReply{
			Aer:  reply,
			Peer: peer,
		}
	}
}
