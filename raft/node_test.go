package raft

import (
	"testing"
)


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
	MakeLeader(n)

	if n.State != LEADER {
		t.Errorf("Wrong status: expected LEADER, got %v\n", n.State)
	}
	if len(n.Log) != 2 {
		t.Errorf("Wrong log length: expected %v, got %v\n", 2, len(n.Log))
	}
    n.Kill()
}

