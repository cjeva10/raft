package main

import (
	"log"
	"os"
	"strconv"

	"github.com/cjeva10/raft/raft"
)

func main() {
    port, err := strconv.Atoi(os.Args[1])
    if err != nil {
        log.Fatal(err)
    }

	raft.Start(port)

	for {
	}
}
