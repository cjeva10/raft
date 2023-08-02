package raft

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
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

	State NodeStates

	// misc
	Pulses int
	mu     sync.Mutex

	PeerList []int

	Id int // id = port
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
	}

	n.CurrentTerm = args.Term
	n.VotedFor = 0
	reply.Success = true
	// fmt.Printf("Term: %v, Received good heartbeat from %v, resetting timer\n", n.CurrentTerm, args.LeaderId)
    fmt.Printf("Current Log: %v, CommitIndex: %v\n", n.Log, n.CommitIndex)

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
		fmt.Printf("granting vote to %v\n", args.CandidateId)
		n.VotedFor = args.CandidateId
		reply.VoteGranted = true

		go n.pulseCheck()
	}

	return nil
}

func (n *Node) server(port int) {
	rpc.Register(n)
	rpc.HandleHTTP()

	portname := strconv.Itoa(port)

	l, err := net.Listen("tcp", ":"+portname)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	go http.Serve(l, nil)
}

func Start(port int) *Node {
	// init Node
	n := Node{}

	// create peer list
	peers := []int{1234, 1235, 1236, 1237, 1238}
	for i, peer := range peers {
		if peer == port {
			copy(peers[i:], peers[i+1:])
			peers = peers[:len(peers)-1]
		}
	}
	fmt.Println(peers)
	n.PeerList = peers

	// read persistent state from storage
	// for now assume it's a fresh boot every time
	n.CurrentTerm = 0
	n.VotedFor = 0
	n.Log = []LogEntry{{
		Command: "",
		Term:    0,
	}}

	n.NextIndex = make(map[int]int)
	n.MatchIndex = make(map[int]int)

	// initialize volatile state
	n.CommitIndex = 0
	n.LastApplied = 0
	n.mu = sync.Mutex{}
	n.CommitCond = *sync.NewCond(&n.mu)

	// start as a follower
	n.State = FOLLOWER

	// the id should be the port we're listening on
	n.Id = port

	// start server
	n.server(port)

	// start the heartbeat
	go n.pulseCheck()

	return &n
}
