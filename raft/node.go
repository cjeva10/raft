package raft

// definition of an raft node

import (
    "sync"
)

type Node struct {
	// persistent
	CurrentTerm int
	VotedFor    int
	Log         []LogEntry

	// volatile (all servers)
	CommitIndex int
	LastApplied int

	// volatile (leaders)
	NextIndex  map[int]int
	MatchIndex map[int]int
	CommitCond sync.Cond

	// current state of this node
	State NodeStates

	// misc
	Pulses int
	mu     sync.Mutex

	// list of peers
	PeerList []int

	// Id of this node (also port that we're listening on)
	Id int

	// Id of current leader
	LeaderId int

    // for testing
    Testing bool
    Peers map[int]*Node
}

type LogEntry struct {
	Term    int
	Command string
}

type NodeStates int

const (
	FOLLOWER  = 0
	CANDIDATE = 1
	LEADER    = 2
)

