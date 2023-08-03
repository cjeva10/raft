package raft

import (
	"testing"
	"time"
)

// Make sure leader can always replicate logs in followers
// See Figure 7 (a-f) from the Raft paper for all the scenarios
func TestLeaderAddEntriesNoClientRequests(t *testing.T) {
	leaderLog := []LogEntry{
		{Term: 0, Command: ""},
		{Term: 1, Command: "1"},
		{Term: 1, Command: "2"},
		{Term: 4, Command: "3"},
		{Term: 4, Command: "4"},
		{Term: 5, Command: "5"},
		{Term: 5, Command: "6"},
		{Term: 6, Command: "7"},
		{Term: 6, Command: "8"},
		{Term: 6, Command: "9"},
	}

	tests := []struct {
		followerLog []LogEntry
	}{
		{ // 0
			leaderLog,
		},
		{ // a
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
				{Term: 4, Command: "4"},
				{Term: 5, Command: "5"},
				{Term: 5, Command: "6"},
				{Term: 6, Command: "7"},
				{Term: 6, Command: "8"},
			},
		},
		{ // b
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
			},
		},
		{ // c
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
				{Term: 4, Command: "4"},
				{Term: 5, Command: "5"},
				{Term: 5, Command: "6"},
				{Term: 6, Command: "7"},
				{Term: 6, Command: "8"},
				{Term: 6, Command: "9"},
				{Term: 6, Command: "10"},
			},
		},
		{ // d
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
				{Term: 4, Command: "4"},
				{Term: 5, Command: "5"},
				{Term: 5, Command: "6"},
				{Term: 6, Command: "7"},
				{Term: 6, Command: "8"},
				{Term: 6, Command: "9"},
				{Term: 7, Command: "10"},
				{Term: 7, Command: "11"},
			},
		},
		{ // e
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
				{Term: 4, Command: "4"},
				{Term: 4, Command: "5"},
				{Term: 4, Command: "6"},
			},
		},
		{ // f
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 2, Command: "3"},
				{Term: 2, Command: "4"},
				{Term: 2, Command: "5"},
				{Term: 3, Command: "6"},
				{Term: 3, Command: "7"},
				{Term: 3, Command: "8"},
				{Term: 3, Command: "9"},
				{Term: 3, Command: "10"},
			},
		},
	}

	for _, tt := range tests {
		runLeaderTest(t, leaderLog, tt.followerLog, nil)
	}
}

// test replication when new logs are added
func TestLeaderAddEntriesWithClientRequests(t *testing.T) {
	leaderLog := []LogEntry{
		{Term: 0, Command: ""},
		{Term: 1, Command: "1"},
		{Term: 1, Command: "2"},
		{Term: 4, Command: "3"},
		{Term: 4, Command: "4"},
		{Term: 5, Command: "5"},
		{Term: 5, Command: "6"},
		{Term: 6, Command: "7"},
		{Term: 6, Command: "8"},
		{Term: 6, Command: "9"},
	}

	tests := []struct {
		followerLog []LogEntry
		requests    []ClientRequestArgs
	}{
		{ // 0
			leaderLog,
			[]ClientRequestArgs{
				{"1"},
                {"2"},
            },
		},
		{ // a
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
				{Term: 4, Command: "4"},
				{Term: 5, Command: "5"},
				{Term: 5, Command: "6"},
				{Term: 6, Command: "7"},
				{Term: 6, Command: "8"},
			},
			[]ClientRequestArgs{
				{"1"},
                {"2"},
            },
		},
		{ // b
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
			},
			[]ClientRequestArgs{
				{"1"},
                {"2"},
            },
		},
		{ // c
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
				{Term: 4, Command: "4"},
				{Term: 5, Command: "5"},
				{Term: 5, Command: "6"},
				{Term: 6, Command: "7"},
				{Term: 6, Command: "8"},
				{Term: 6, Command: "9"},
				{Term: 6, Command: "10"},
			},
			[]ClientRequestArgs{
				{"1"},
                {"2"},
            },
		},
		{ // d
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
				{Term: 4, Command: "4"},
				{Term: 5, Command: "5"},
				{Term: 5, Command: "6"},
				{Term: 6, Command: "7"},
				{Term: 6, Command: "8"},
				{Term: 6, Command: "9"},
				{Term: 7, Command: "10"},
				{Term: 7, Command: "11"},
			},
			[]ClientRequestArgs{
				{"1"},
                {"2"},
            },
		},
		{ // e
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 4, Command: "3"},
				{Term: 4, Command: "4"},
				{Term: 4, Command: "5"},
				{Term: 4, Command: "6"},
			},
			[]ClientRequestArgs{
				{"1"},
                {"2"},
            },
		},
		{ // f
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 1, Command: "2"},
				{Term: 2, Command: "3"},
				{Term: 2, Command: "4"},
				{Term: 2, Command: "5"},
				{Term: 3, Command: "6"},
				{Term: 3, Command: "7"},
				{Term: 3, Command: "8"},
				{Term: 3, Command: "9"},
				{Term: 3, Command: "10"},
			},
			[]ClientRequestArgs{
				{"1"},
                {"2"},
            },
		},
	}

	for _, tt := range tests {
		runLeaderTest(t, leaderLog, tt.followerLog, tt.requests)
	}
}

func runLeaderTest(t *testing.T, leaderLog []LogEntry, followerLog []LogEntry, requests []ClientRequestArgs) {
	term := 8

	nodes := SetupTestNodes(2)

	log := leaderLog
	nodes[0].Log = log

	nodes[0].CurrentTerm = term
	nodes[0].VotedFor = nodes[0].Id
	nodes[0].State = LEADER

	log2 := followerLog
	nodes[1].Log = log2
	nodes[1].CurrentTerm = term - 1
	nodes[1].State = FOLLOWER

	go nodes[0].Start(true)
	go nodes[0].leader()

	go nodes[1].Start(true)

	for _, req := range requests {
		reply := ClientRequestReply{}
		go nodes[0].ClientRequest(&req, &reply)
	}

	time.Sleep(2 * PULSETIME * time.Millisecond)

	log3 := nodes[1].Log

    if len(log) + len(requests) != len(nodes[0].Log) {
        t.Errorf("Leader failed to add requests\n")
    } 

	// make sure that all the leader's entries are replicated on the follower's
	for i, entry := range nodes[0].Log {
		if entry.Term != log3[i].Term {
			t.Errorf("Failed to replicate log entry: expected log[%v]=%v, got log[%v]=%v", i, log[i], i, entry)
		}
	}

	KillTestNodes(nodes)
}
