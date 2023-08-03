package tests

import (
    "fmt"
    "log"

    "github.com/cjeva10/raft/raft"
)

func SetupTestNodes(count int) []*raft.Node {
	nodes := []*raft.Node{}

	if count > 5 || count <= 0 {
		log.Fatalf("Invalid count to setup\n")
	}
	for i := 0; i < count; i++ {
		nodes = append(nodes, raft.SetupNode(1234+i))
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

func StartTestNodes(nodes []*raft.Node) {
	for _, n := range nodes {
		go n.Start(true)
	}
}

func KillTestNodes(nodes[]*raft.Node) {
    for _, n := range nodes {
        n.Kill()
    }
}
