# Raft

## Overview

Implementing the Raft consensus algorithm in Go.

For now, nodes communicate over tcp on a single host. 

From top-level directory, run `go run main.go [port-number]` to start the node listening on the specified port.
Right now, 5 fixed nodes are configured in the peer list, using ports 1234-1238. So at least 3 are needed to elect a leader and field commands.

To run the client, run `go run client/client.go`. 
Send RPC requests to the cluster on stdin (eg `1 1234\n` sends the command `1` to port `1234`)

## Status

 - Implemented the basic leader election, tested with 5 nodes
 - Implemented appending logs
 - First draft AppendEntries and RequestVote RPCs done
 - Simple KV store implemented as basic application
 - Mock RPC calls for testing
 - Unit tests on RPC handlers, elections, kv store, leader functions

 - Next Steps: 
     - Apply Committed logs to kv store
     - Persist state on disk.

 - Known issues:
     - Sometimes leader elections fail in tests
         - looks like too many threads competing for CPU time means election timeouts happen before RPCs can be handled.
         - shouldn't happen when using actual sockets

## testing

Need to set up basic infrastructure to create specific custom specific scenarios for testing.
Not sure how to do that yet.

Some scenarios that need to be tested

 - Elections:
    - Test base election case (nodes starting from nothing)
    - Test that only most up to date server becomes leader
    - Test that heartbeat is reset when vote granted
    - Test that lower term request vote always rejected
    - Test that if already voted, do not vote again in the same term
    - Ensure that if term changes, VotedFor always becomes null
    - Test voting for self

 - Leader Election:
    - Test all cases from Figure 7 of raft paper:
        - Spawn follower nodes with all the scenarios a-f and ensure that leader successfully replicates its log.

 - Log replication:
    - Test that if majority of servers are running, client requests always go through
    - Test sending RPC to wrong server (make sure we are redirected to the leader)
    - Test timeout - i.e. if the service is not responding make sure that we can detect that properly and mark the request as failed
