package raft

import (
	"testing"
	"time"
)

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

