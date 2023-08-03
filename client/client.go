package main

import (
	"errors"
	"fmt"
	"net/rpc"
	"os"
	"strconv"

	"github.com/cjeva10/raft/raft"
)

func Start(port int) bool {
	return false
}

func call(peer int, rpcname string, args interface{}, reply interface{}) bool {
	peername := strconv.Itoa(peer+1230)

	c, err := rpc.DialHTTP("tcp", "localhost:"+peername)
	if err != nil {
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	return false
}

func callClientRequest(val int, port int) error {
	args := raft.ClientRequestArgs{
		Command: strconv.Itoa(val),
	}
	reply := raft.ClientRequestReply{}

	ok := call(port, "Node.ClientRequest", &args, &reply)
	if ok {
		fmt.Printf("Received reply: %v\n", reply)
	} else {
		return errors.New("Bad reply from server")
	}

	return nil
}

// receive commands from stdin and send requests to the leader
func main() {
	for {
		var val int
		var port int

		_, err := fmt.Scanf("%d %d\n", &val, &port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
			continue
		}

		err = callClientRequest(val, port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}
