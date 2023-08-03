package raft

import (
    "fmt"
	"log"
	"testing"
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
