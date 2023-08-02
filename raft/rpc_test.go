package raft

import (
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
}

func TestElectLeader(t *testing.T) {
    // setup just three nodes
    n1 := SetupNode(1234)
    n2 := SetupNode(1235)
    n3 := SetupNode(1236)
    
    n1.Peers[1236] = n3
    n1.Peers[1235] = n2 
    n2.Peers[1234] = n1 
    n2.Peers[1236] = n3
    n3.Peers[1234] = n1 
    n3.Peers[1235] = n2

    n1.Start(true)
	n2.Start(true)
	n3.Start(true)

	// wait for more than a pulse and make sure that we have a leader
	time.Sleep(1.5 * PULSETIME * time.Millisecond)

	foundLeader := false
	if n1.State == LEADER {
		foundLeader = true
	}
	if n2.State == LEADER {
		foundLeader = true
	}
	if n3.State == LEADER {
		foundLeader = true
	}

	if !foundLeader {
		t.Errorf("Failed to establish leader!\n")
	}
}
