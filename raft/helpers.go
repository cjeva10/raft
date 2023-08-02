package raft

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

// min heartbeat time in milliseconds
const PULSETIME = 50 

// helper function to call RPC methods on a peer
func (n *Node) call(peer int, rpcname string, args interface{}, reply interface{}) bool {
	if !n.Testing {
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
	} else {
        return n.MockCall(peer, rpcname, args, reply)
    }
}

// make a mock call to a test peer if we are testing
func (n *Node) MockCall(peer int, rpcname string, args interface{}, reply interface{}) bool {
    if !n.Testing {
        log.Fatalf("Can only use mockCall when in testing\n")
    }

	peerNode := n.Peers[peer]
    if peerNode == nil {
        return false
    }

	switch rpcname {
	case "Node.RequestVote":
		err := peerNode.RequestVote(args.(*RequestVoteArgs), reply.(*RequestVoteReply))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad reply to RequestVote, err: %v\n", err)
			return false
		}
	case "Node.AppendEntries":
		err := peerNode.AppendEntries(args.(*AppendEntriesArgs), reply.(*AppendEntriesReply))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad reply to AppendEntries, err: %v\n", err)
			return false
		}
	case "Node.ClientRequest":
		err := peerNode.ClientRequest(args.(*ClientRequestArgs), reply.(*ClientRequestReply))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad reply to ClientRequest, err: %v\n", err)
			return false
		}
	default:
		log.Fatalf("Unknown rpc method: %v", rpcname)
		return false
	}

	return true
}
// wait for a random amount of time and then request votes
// for debugging the min wait time is 2 seconds
func (n *Node) pulseCheck() {
	// kill all existing pulses (by incrementing counter) and record current total
	n.mu.Lock()
	n.Pulses++
	currPulses := n.Pulses
	n.mu.Unlock()

	// wait
	delay := PULSETIME + rand.Intn(PULSETIME/4)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// if we didn't receive a pulse, call an election
	n.mu.Lock()
	if currPulses == n.Pulses {
		go n.callElection() // do this on another thread to release the lock
	}
	n.mu.Unlock()
}

func (n *Node) becomeFollower(updateTerm int) {
	n.CurrentTerm = updateTerm
	n.VotedFor = 0 // reset vote whenever we update our term
	n.State = FOLLOWER
	go n.pulseCheck() // restart election timer
}

func (n *Node) checkLastApplied() {
	if n.CommitIndex > n.LastApplied {
		n.LastApplied++
		go n.applyToStateMachine(n.Log[n.LastApplied])
	}
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

func (n *Node) applyToStateMachine(entry LogEntry) {
	fmt.Printf("Applying Log to Machine: %v\n", entry)

	// todo
	// probably we want to have a channel or queue so that the state machine
	// requests are always processed correctly in order
}

func SetupNode(port int) *Node {
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

    // for testing
    n.Peers = make(map[int]*Node)

	return &n
}
