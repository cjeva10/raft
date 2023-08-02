# Raft

## Overview

Implementing the Raft consensus algorithm in Go.

For now, nodes communicate over tcp on a single host. 

From top-level directory, run `go run main.go [port-number]` to start the node listening on the specified port.

## Status

 - Implemented the basic leader election, tested with 5 nodes
 - Next Steps: 
 - Appending logs (create a simple service)
 - more detailed RPC handling for AppendEntries, 
 - Unit testing
 - store persistent state on disk.

## testing

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
