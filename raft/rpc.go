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

	// if the sender's term is less than our term, automatically reject this request
	if args.Term < n.CurrentTerm {
		reply.Success = false
		return nil
	}

	// log doesn't contain an entry at PrevLogIndex whose term mathces PrevLogTerm
	if len(n.Log)-1 < args.PrevLogIndex || n.Log[args.PrevLogIndex].Term != args.PrevLogTerm {
		reply.Success = false
		return nil
	}

	n.CurrentTerm = args.Term
	n.VotedFor = 0
	reply.Success = true
	fmt.Printf("Term: %v, Received good heartbeat from %v, resetting timer\n", n.CurrentTerm, args.LeaderId)

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

	reply.Term = n.CurrentTerm

	fmt.Printf("Received vote request from %v: term = %v\n", args.CandidateId, args.Term)

	if args.Term > n.CurrentTerm {
		n.becomeFollower(args.Term)
	}

	if (n.VotedFor == 0 || n.VotedFor == args.CandidateId) && args.Term >= n.CurrentTerm {
		fmt.Printf("granting vote to %v\n", args.CandidateId)
		n.VotedFor = args.CandidateId
		reply.VoteGranted = true

		go n.pulseCheck()
	}

	return nil
}

type LogEntry struct {
	Term    int
	Command string
}

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

type NodeStates int

const (
	FOLLOWER  = 0
	CANDIDATE = 1
	LEADER    = 2
)

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
