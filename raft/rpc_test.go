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

    n := Start(port)
    if n.Id != port {
        t.Errorf("Wrong id: expected %v, got %v\n", port, n.Id)
    }
}

func TestMakeLeader(t *testing.T) {
    port := 1234
    n := setupNode(port)
    makeLeader(n)

    if n.State != LEADER {
        t.Errorf("Wrong status: expected LEADER, got %v\n", n.State)
    }
    if len(n.Log) != 2 {
        t.Errorf("Wrong log length: expected %v, got %v\n", 2, len(n.Log))
    }
}

func TestElectLeader(t *testing.T) {
    // setup three nodes
    n1 := Start(1234)
    n2 := Start(1235)
    n3 := Start(1236) 

    // wait for two pulses and make sure that we have a leader
    time.Sleep(2 * PULSETIME * time.Millisecond)

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
