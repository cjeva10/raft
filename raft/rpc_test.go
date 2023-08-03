package raft

import (
    "fmt"
	"log"
	"testing"
	"time"
)

func makeLeader(n *Node) {
	n.Log = []LogEntry{
		{
			Command: "",
			Term:    0,
		},
		{
			Command: "1",
			Term:    1,
		},
	}

	n.State = LEADER
}

func TestStartEmptyRPC(t *testing.T) {
	port := 1234

	n := SetupNode(port)
	n.Start(true)
	if n.Id != port {
		t.Errorf("Wrong id: expected %v, got %v\n", port, n.Id)
	}
    n.Kill()
}

func TestMakeLeader(t *testing.T) {
	port := 1234
	n := SetupNode(port)
	makeLeader(n)

	if n.State != LEADER {
		t.Errorf("Wrong status: expected LEADER, got %v\n", n.State)
	}
	if len(n.Log) != 2 {
		t.Errorf("Wrong log length: expected %v, got %v\n", 2, len(n.Log))
	}
    n.Kill()
}

func SetupTestNodes(count int) []*Node {
	nodes := []*Node{}

	if count > 5 || count <= 0 {
		log.Fatalf("Invalid count to setup\n")
	}
	for i := 0; i < count; i++ {
		nodes = append(nodes, SetupNode(1234+i))
	}

    fmt.Printf("count = %v\n", count)
    fmt.Printf("Nodes = %v\n", nodes)

	for i, n := range nodes {
        for _, n2 := range nodes[i:] {
			if n.Id != n2.Id {
				n.Peers[n2.Id] = n2
			}
		}
	}

	return nodes
}

func StartTestNodes(nodes []*Node) {
	for _, n := range nodes {
		n.Start(true)
	}
}

func KillTestNodes(nodes[]*Node) {
    for _, n := range nodes {
        n.Kill()
    }
}

func TestElectLeaderFiveNodes(t *testing.T) {
	nodes := SetupTestNodes(5)

	StartTestNodes(nodes)

	// wait for more than a pulse and make sure that we have a leader
	time.Sleep(5 * PULSETIME * time.Millisecond)

	foundLeader := 0
	for _, n := range nodes {
		if n.State == LEADER {
			foundLeader++
		}
	}

	if foundLeader < 1 {
		t.Errorf("Failed to establish leader!\n")
	}
	if foundLeader > 1 {
		t.Errorf("More than one leader\n")
	}

    KillTestNodes(nodes)
}

func TestRequestVoteOnFollower(t *testing.T) {
	tests := []struct {
		args  RequestVoteArgs
		reply RequestVoteReply
	}{
		// lower term -> false
		{
			RequestVoteArgs{
				Term:         4,
				CandidateId:  1235,
				LastLogIndex: 1,
				LastLogTerm:  1,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: false,
			},
		},
		// higher LastLogTerm -> true
		{
			RequestVoteArgs{
				Term:         5,
				CandidateId:  1235,
				LastLogIndex: 3,
				LastLogTerm:  3,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: true,
			},
		},
		// lower LastLogIndex -> false
		{
			RequestVoteArgs{
				Term:         5,
				CandidateId:  1235,
				LastLogIndex: 1,
				LastLogTerm:  2,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: false,
			},
		},
		// greater/equal LastLogIndex -> true
		{
			RequestVoteArgs{
				Term:         5,
				CandidateId:  1235,
				LastLogIndex: 2,
				LastLogTerm:  2,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: true,
			},
		},
	}

	for _, tt := range tests {
		n := SetupNode(1234)
		n.Log = append(n.Log, LogEntry{Term: 1, Command: "1"})
		n.Log = append(n.Log, LogEntry{Term: 2, Command: "2"})
		n.CurrentTerm = 5
		n.VotedFor = 0

		actualReply := RequestVoteReply{}
		err := n.RequestVote(&tt.args, &actualReply)
		if err != nil {
			t.Errorf("Error when requesting vote: %v\n", err)
		}

		if tt.reply.Term != actualReply.Term {
			t.Errorf("Mismatching Term: expected %v, got %v\n", tt.reply.Term, actualReply.Term)
		}
		if tt.reply.VoteGranted != actualReply.VoteGranted {
			t.Errorf("Mismatching Vote: expected %v, got %v\n", tt.reply.VoteGranted, actualReply.VoteGranted)
		}
        n.Kill()
	}
}

func TestRequestVoteOnCandidate(t *testing.T) {
	tests := []struct {
		args  RequestVoteArgs
		reply RequestVoteReply
	}{
		// lower term -> false
		{
			RequestVoteArgs{
				Term:         4,
				CandidateId:  1235,
				LastLogIndex: 1,
				LastLogTerm:  1,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: false,
			},
		},
		// higher term, same logs -> true
		{
			RequestVoteArgs{
				Term:         6,
				CandidateId:  1235,
				LastLogIndex: 2,
				LastLogTerm:  2,
			},
			RequestVoteReply{
				Term:        6,
				VoteGranted: true,
			},
		},
		// higher LastLogTerm -> false (already voted)
		{
			RequestVoteArgs{
				Term:         5,
				CandidateId:  1235,
				LastLogIndex: 10,
				LastLogTerm:  4,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: false,
			},
		},
	}

	for _, tt := range tests {
		n := SetupNode(1234)

		n.Log = append(n.Log, LogEntry{Term: 1, Command: "1"})
		n.Log = append(n.Log, LogEntry{Term: 2, Command: "2"})
		n.CurrentTerm = 5
		n.VotedFor = 1234
		n.State = CANDIDATE

		actualReply := RequestVoteReply{}
		err := n.RequestVote(&tt.args, &actualReply)
		if err != nil {
			t.Errorf("Error when requesting vote: %v\n", err)
		}

		if tt.reply.Term != actualReply.Term {
			t.Errorf("Mismatching Term: expected %v, got %v\n", tt.reply.Term, actualReply.Term)
		}
		if tt.reply.VoteGranted != actualReply.VoteGranted {
			t.Errorf("Mismatching Vote: expected %v, got %v\n", tt.reply.VoteGranted, actualReply.VoteGranted)
		}
	}
}

