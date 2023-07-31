package main

import (
	"log"
	"os"
	"strconv"

	"github.com/cjeva10/raft/rpc"
)

func main() {
    port, err := strconv.Atoi(os.Args[1])
    if err != nil {
        log.Fatal(err)
    }

	rpc.Start(port)

	for {
	}
}
