package raft

// implementation of RPC handles and RPC initalization

import (
	"fmt"
)

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (n *Node) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	reply.Term = n.CurrentTerm

	// 1. if the sender's term is less than our term, automatically reject this request
	if args.Term < n.CurrentTerm {
		reply.Success = false
		return nil
	}

	// 2. log doesn't contain an entry at PrevLogIndex whose term matches PrevLogTerm
	if len(n.Log)-1 < args.PrevLogIndex || n.Log[args.PrevLogIndex].Term != args.PrevLogTerm {
		fmt.Printf("Term: %v, inconsistent log\n", n.CurrentTerm)

		reply.Success = false
		return nil
	}

	// 3. if an existing entry conflicts with a new one (same index but different terms)
	// delete the existing entry and all that follow it
	idx := args.PrevLogIndex + 1
	for _, entry := range args.Entries {
		// idx out of bounds means that all entries are new
		if idx > len(n.Log)-1 {
			break
		}
		if entry.Term != n.Log[idx].Term {
			n.Log = n.Log[:idx]
			break
		}
		idx++
	}

	// 4. append any new entries not in the log
	idx = args.PrevLogIndex + 1
	for i, entry := range args.Entries {
		if idx > len(n.Log)-1 {
			n.Log = append(n.Log, args.Entries[i:]...)
		}

		if entry.Term != n.Log[idx].Term {
			n.Log[idx] = entry
		}

		idx++
	}

	// receiver implementation 5.
	if args.LeaderCommit > n.CommitIndex {
		idxLastNewEntry := len(n.Log) - 1
		if idxLastNewEntry < args.LeaderCommit {
			n.CommitIndex = idxLastNewEntry
		} else {
			n.CommitIndex = args.LeaderCommit
		}

		n.checkLastApplied()
	}

	n.CurrentTerm = args.Term
	n.VotedFor = 0
	n.LeaderId = args.LeaderId
	reply.Success = true
	// fmt.Printf("Term: %v, Received good heartbeat from %v, resetting timer\n", n.CurrentTerm, args.LeaderId)
	fmt.Printf("Successful heartbeat: Current Log: %v, CommitIndex: %v, CurrentLeader: %v\n", n.Log, n.CommitIndex, n.LeaderId)

	// reset election timer
	go n.pulseCheck()

	return nil
}

type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func (n *Node) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term > n.CurrentTerm {
		n.becomeFollower(args.Term)
	}

	reply.Term = n.CurrentTerm

	fmt.Printf("Received vote request from %v: term = %v\n", args.CandidateId, args.Term)

	if args.Term < n.CurrentTerm {
		reply.VoteGranted = false
		return nil
	}

	if n.VotedFor == 0 || n.VotedFor == args.CandidateId {
		ourLastIndex := len(n.Log) - 1
		ourLastTerm := n.Log[ourLastIndex].Term
		if args.LastLogTerm > ourLastTerm {
			fmt.Printf("granting vote to %v\n", args.CandidateId)
			n.VotedFor = args.CandidateId
			reply.VoteGranted = true
		} else if args.LastLogTerm == ourLastTerm {
			if args.LastLogIndex >= ourLastIndex {
				fmt.Printf("granting vote to %v\n", args.CandidateId)
				n.VotedFor = args.CandidateId
				reply.VoteGranted = true
			}
		}
		go n.pulseCheck()
	}

	return nil
}

func Start(port int) *Node {
	n := setupNode(port)

	// start server
	n.server(port)

	// start the heartbeat
	go n.pulseCheck()

	return n
}
