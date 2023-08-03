package tests

import (
	"testing"
	"time"

    "github.com/cjeva10/raft/raft"
)

func TestElectLeaderFiveNodes(t *testing.T) {
	nodes := raft.SetupTestNodes(5)

	raft.StartTestNodes(nodes)
	defer raft.KillTestNodes(nodes)

	// wait for more than a pulse and make sure that we have a leader
	time.Sleep(10 * raft.PULSETIME * time.Millisecond)

	foundLeader := 0
	for _, n := range nodes {
		if n.State == raft.LEADER {
			foundLeader++
		}
	}

	if foundLeader < 1 {
		t.Errorf("Failed to establish leader!\n")
	}
	if foundLeader > 1 {
		t.Errorf("More than one leader\n")
	}
}
