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

	fmt.Printf("%v: AppendEntries: received from %v", n.Id-1230, args.LeaderId)

	// our Term is out of date
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.VotedFor = 0
		n.State = FOLLOWER
	}

	reply.Term = n.CurrentTerm

	// 1. if the sender's term is less than our term, automatically reject this request
	if args.Term < n.CurrentTerm {
		fmt.Printf("%v: AppendEntries: leader term less than ours: got %v, have %v\n", n.Id-1230, args.Term, n.CurrentTerm)
		reply.Success = false
		return nil
	}

	// 2. log doesn't contain an entry at PrevLogIndex whose term matches PrevLogTerm
	if len(n.Log)-1 < args.PrevLogIndex {
		fmt.Printf("%v: AppendEntries from %v: inconsistent log: ourIndex %v, args.Index %v\n",
			n.Id-1230,
			n.LeaderId,
			len(n.Log)-1,
			args.PrevLogIndex,
		)

		reply.Success = false
		return nil
	}

	if n.Log[args.PrevLogIndex].Term != args.PrevLogTerm {
		fmt.Printf("%v: AppendEntries from %v: inconsistent log: ourPrevTerm %v, args.PrevLogTerm %v\n",
			n.Id-1230,
			n.LeaderId,
			n.Log[args.PrevLogIndex].Term,
			args.PrevLogTerm,
		)

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
    for i := 0; i < len(args.Entries); i++ {
		if idx + i > len(n.Log)-1 {
			n.Log = append(n.Log, args.Entries[i:]...)
            break
		}
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

	n.VotedFor = args.LeaderId
	n.LeaderId = args.LeaderId
	reply.Success = true
	fmt.Printf("%v: AppendEntries Successful: Current Log: %v, CommitIndex: %v, CurrentLeader: %v\n", n.Id-1230, n.Log, n.CommitIndex, n.LeaderId)

	// reset election timer
	go n.resetTimer()

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

	// our Term is out of date
	// NOTE: don't reset timer because we have not granted a vote
	// i.e candidate might have higher term but outdated log
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.VotedFor = 0
		n.State = FOLLOWER
	}

	reply.Term = n.CurrentTerm

	fmt.Printf("%v: Received vote request from %v, term = %v\n", n.Id, args.CandidateId, args.Term)

	// sender has outdated term
	if args.Term < n.CurrentTerm {
		fmt.Printf("%v: Rejected vote, outdated term: got %v, have %v\n", n.Id-1230, args.Term, n.CurrentTerm)
		reply.VoteGranted = false
		return nil
	}

	// if we haven't voted, and candidate log is at least as up to date as ours
	if n.VotedFor == 0 || n.VotedFor == args.CandidateId {
		ourLastIndex := len(n.Log) - 1
		ourLastTerm := n.Log[ourLastIndex].Term

		if args.LastLogTerm > ourLastTerm {

			fmt.Printf("%v: granting vote to %v\n", n.Id, args.CandidateId)
			n.VotedFor = args.CandidateId
			reply.VoteGranted = true
			go n.resetTimer()

		} else if args.LastLogTerm == ourLastTerm {

			if args.LastLogIndex >= ourLastIndex {

				fmt.Printf("%v: granting vote to %v\n", n.Id, args.CandidateId)
				n.VotedFor = args.CandidateId
				reply.VoteGranted = true
				go n.resetTimer()
			}
		}
	} else {
		ourLastIndex := len(n.Log) - 1
		ourLastTerm := n.Log[ourLastIndex].Term

		reply.VoteGranted = false
		fmt.Printf("%v: Rejected vote, Log outdated: got Term: %v, Index: %v, have Term: %v, Index: %v\n", n.Id-1230, args.LastLogTerm, args.LastLogIndex, ourLastTerm, ourLastIndex)
	}

	return nil
}

func (n *Node) Start(testing bool) *Node {
	// start server
	if testing {
		n.Testing = true
	} else {
		n.server(n.Id)
	}

	// start the heartbeat
	go n.resetTimer()

	return n
}
