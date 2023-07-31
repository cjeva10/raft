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
