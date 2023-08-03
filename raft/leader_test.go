package raft

import (
	"testing"
	"time"
)

// Make sure leader can always replicate logs in followers
// See Figure 7 (a-f) from the Raft paper for all the scenarios
func TestLeaderAddEntries(t *testing.T) {
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
		runLeaderTest(t, leaderLog, tt.followerLog)
	}
}

func runLeaderTest(t *testing.T, leaderLog []LogEntry, followerLog []LogEntry) {
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

	time.Sleep(2 * PULSETIME * time.Millisecond)

	log3 := nodes[1].Log

    // make sure that all the leader's entries are replicated on the follower's
	for i, entry := range log {
		if entry.Term != log3[i].Term {
			t.Errorf("Failed to replicate log entry: expected log[%v]=%v, got log[%v]=%v", i, log[i], i, entry)
		}
	}

	KillTestNodes(nodes)

}

