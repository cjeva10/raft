package raft

import (
	"testing"
)

func TestAppendEntriesLogAppend(t *testing.T) {
	tests := []struct {
		args        AppendEntriesArgs
		expectedLog []LogEntry
	}{
		// simple append
		{
			AppendEntriesArgs{
				Term:     5,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 5, Command: "3"},
				},
				PrevLogIndex: 2,
				PrevLogTerm:  2,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 2, Command: "2"},
				{Term: 5, Command: "3"},
			},
		},
		// append but higher term
		{
			AppendEntriesArgs{
				Term:     6,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 6, Command: "3"},
				},
				PrevLogIndex: 2,
				PrevLogTerm:  2,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 2, Command: "2"},
				{Term: 6, Command: "3"},
			},
		},
		// append but lower term
		{
			AppendEntriesArgs{
				Term:     4,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 4, Command: "3"},
				},
				PrevLogIndex: 2,
				PrevLogTerm:  2,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 2, Command: "2"},
			},
		},
		// replace last entry
		{
			AppendEntriesArgs{
				Term:     6,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 6, Command: "3"},
				},
				PrevLogIndex: 1,
				PrevLogTerm:  1,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 6, Command: "3"},
			},
		},
		// replace earlier entry
		{
			AppendEntriesArgs{
				Term:     6,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 3, Command: "3"},
				},
				PrevLogIndex: 0,
				PrevLogTerm:  0,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 3, Command: "3"},
			},
		},
		// leader term lower than ours
		{
			AppendEntriesArgs{
				Term:     4,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 4, Command: "3"},
				},
				PrevLogIndex: 3,
				PrevLogTerm:  3,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 2, Command: "2"},
			},
		},
		// inconsistent log -> PrevLogIndex too large
		{
			AppendEntriesArgs{
				Term:     6,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 3, Command: "3"},
				},
				PrevLogIndex: 3,
				PrevLogTerm:  3,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 2, Command: "2"},
			},
		},
		// inconsistent log -> PrevLogTerm doesn't match
		{
			AppendEntriesArgs{
				Term:     6,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 6, Command: "3"},
				},
				PrevLogIndex: 2,
				PrevLogTerm:  5,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 2, Command: "2"},
			},
		},
		// append multiple entries at the back of the log
		{
			AppendEntriesArgs{
				Term:     6,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 6, Command: "3"},
					{Term: 6, Command: "3"},
					{Term: 6, Command: "3"},
					{Term: 6, Command: "3"},
				},
				PrevLogIndex: 2,
				PrevLogTerm:  2,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 2, Command: "2"},
				{Term: 6, Command: "3"},
				{Term: 6, Command: "3"},
				{Term: 6, Command: "3"},
				{Term: 6, Command: "3"},
			},
		},
        // overwrite our entries and then add more on the back
		{
			AppendEntriesArgs{
				Term:     6,
				LeaderId: 5,
				Entries: []LogEntry{
					{Term: 6, Command: "3"},
					{Term: 6, Command: "3"},
					{Term: 6, Command: "3"},
					{Term: 6, Command: "3"},
				},
				PrevLogIndex: 0,
				PrevLogTerm:  0,
			},
			[]LogEntry{
				{Term: 0, Command: ""},
				{Term: 1, Command: "1"},
				{Term: 2, Command: "2"},
				{Term: 6, Command: "3"},
				{Term: 6, Command: "3"},
				{Term: 6, Command: "3"},
				{Term: 6, Command: "3"},
			},
		},
	}

	for _, tt := range tests {
		id := 4

		n := SetupNode(id)
		n.Log = []LogEntry{
			{Term: 0, Command: ""},
			{Term: 1, Command: "1"},
			{Term: 2, Command: "2"},
		}
		n.CurrentTerm = 5
		n.VotedFor = 0

		err := n.AppendEntries(&tt.args, &AppendEntriesReply{})
		if err != nil {
			t.Errorf("Error when requesting vote: %v\n", err)
		}
		if len(tt.expectedLog) != len(n.Log) {
			t.Errorf("Incorrect log length: expected %v, got %v\n", len(tt.expectedLog), len(n.Log))
		}

		for i, entry := range tt.expectedLog {
			if entry.Term != n.Log[i].Term {
				t.Errorf("Incorrect log term: expected Term[%v]=%v, got Term[%v]=%v\n", i, entry.Term, i, n.Log[i].Term)
			}
		}
	}
}

func TestAppendEntriesTermUpdate(t *testing.T) {
	term := 5
	log := []LogEntry{
		{Term: 0, Command: ""},
		{Term: 1, Command: "1"},
		{Term: 2, Command: "2"},
	}
	votedFor := 0
	id := 4

	tests := []struct {
		args         AppendEntriesArgs
		expectedTerm int
	}{
		// lower term -> don't update
		{
			AppendEntriesArgs{
				Term:     4,
				LeaderId: 5,
			},
			term,
		},
		// same term -> don't update
		{
			AppendEntriesArgs{
				Term:     5,
				LeaderId: 5,
			},
			term,
		},
		// higher term -> increase term
		{
			AppendEntriesArgs{
				Term:     6,
				LeaderId: 5,
			},
			6,
		},
	}

	for _, tt := range tests {
		n := SetupNode(id)
		n.Log = log
		n.CurrentTerm = term
		n.VotedFor = votedFor

		err := n.AppendEntries(&tt.args, &AppendEntriesReply{})
		if err != nil {
			t.Errorf("Error when requesting vote: %v\n", err)
		}

		if tt.expectedTerm != n.CurrentTerm {
			t.Errorf("Mismatching Term: expected %v, got %v\n", tt.expectedTerm, n.CurrentTerm)
		}
	}
}

func TestAppendEntriesReply(t *testing.T) {
	term := 5
	log := []LogEntry{
		{Term: 0, Command: ""},
		{Term: 1, Command: "1"},
		{Term: 2, Command: "2"},
	}
	votedFor := 0
	id := 4

	tests := []struct {
		args  AppendEntriesArgs
		reply AppendEntriesReply
	}{
		// lower term -> false
		{
			AppendEntriesArgs{
				Term:     4,
				LeaderId: 5,
			},
			AppendEntriesReply{
				Term:    5,
				Success: false,
			},
		},
		// same term -> new leader w/ same log = true
		{
			AppendEntriesArgs{
				Term:         5,
				LeaderId:     5,
				PrevLogTerm:  2,
				PrevLogIndex: 2,
			},
			AppendEntriesReply{
				Term:    5,
				Success: true,
			},
		},
		// same term -> leader w/ longer log = false
		{
			AppendEntriesArgs{
				Term:         5,
				LeaderId:     5,
				PrevLogTerm:  3,
				PrevLogIndex: 3,
			},
			AppendEntriesReply{
				Term:    5,
				Success: false,
			},
		},
		// higher term -> leader w/ longer log = false
		{
			AppendEntriesArgs{
				Term:         6,
				LeaderId:     5,
				PrevLogTerm:  3,
				PrevLogIndex: 3,
			},
			AppendEntriesReply{
				Term:    6,
				Success: false,
			},
		},
		// higher term -> leader w/ inconsistent log = false
		{
			AppendEntriesArgs{
				Term:         6,
				LeaderId:     5,
				PrevLogTerm:  3,
				PrevLogIndex: 2,
			},
			AppendEntriesReply{
				Term:    6,
				Success: false,
			},
		},
		// higher term -> new leader w/ same log = true
		{
			AppendEntriesArgs{
				Term:         6,
				LeaderId:     5,
				PrevLogTerm:  2,
				PrevLogIndex: 2,
			},
			AppendEntriesReply{
				Term:    6,
				Success: true,
			},
		},
	}

	for _, tt := range tests {
		n := SetupNode(id)
		n.Log = log
		n.CurrentTerm = term
		n.VotedFor = votedFor

		actualReply := AppendEntriesReply{}
		err := n.AppendEntries(&tt.args, &actualReply)
		if err != nil {
			t.Errorf("Error when requesting vote: %v\n", err)
		}

		if tt.reply.Term != actualReply.Term {
			t.Errorf("Mismatching Term: expected %v, got %v\n", tt.reply.Term, actualReply.Term)
		}
		if tt.reply.Success != actualReply.Success {
			t.Errorf("Mismatching Success: expected %v, got %v\n", tt.reply.Success, actualReply.Success)
		}
		n.Kill()
	}
}

func TestRequestVoteOnFollower(t *testing.T) {
	tests := []struct {
		args  RequestVoteArgs
		reply RequestVoteReply
	}{
		// lower term -> false
		{
			RequestVoteArgs{
				Term:         4,
				CandidateId:  1235,
				LastLogIndex: 1,
				LastLogTerm:  1,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: false,
			},
		},
		// higher LastLogTerm -> true
		{
			RequestVoteArgs{
				Term:         5,
				CandidateId:  1235,
				LastLogIndex: 3,
				LastLogTerm:  3,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: true,
			},
		},
		// lower LastLogIndex -> false
		{
			RequestVoteArgs{
				Term:         5,
				CandidateId:  1235,
				LastLogIndex: 1,
				LastLogTerm:  2,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: false,
			},
		},
		// greater/equal LastLogIndex -> true
		{
			RequestVoteArgs{
				Term:         5,
				CandidateId:  1235,
				LastLogIndex: 2,
				LastLogTerm:  2,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: true,
			},
		},
	}

	for _, tt := range tests {
		n := SetupNode(1234)
		n.Log = append(n.Log, LogEntry{Term: 1, Command: "1"})
		n.Log = append(n.Log, LogEntry{Term: 2, Command: "2"})
		n.CurrentTerm = 5
		n.VotedFor = 0

		actualReply := RequestVoteReply{}
		err := n.RequestVote(&tt.args, &actualReply)
		if err != nil {
			t.Errorf("Error when requesting vote: %v\n", err)
		}

		if tt.reply.Term != actualReply.Term {
			t.Errorf("Mismatching Term: expected %v, got %v\n", tt.reply.Term, actualReply.Term)
		}
		if tt.reply.VoteGranted != actualReply.VoteGranted {
			t.Errorf("Mismatching Vote: expected %v, got %v\n", tt.reply.VoteGranted, actualReply.VoteGranted)
		}
		n.Kill()
	}
}

func TestRequestVoteOnCandidate(t *testing.T) {
	tests := []struct {
		args  RequestVoteArgs
		reply RequestVoteReply
	}{
		// lower term -> false
		{
			RequestVoteArgs{
				Term:         4,
				CandidateId:  1235,
				LastLogIndex: 1,
				LastLogTerm:  1,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: false,
			},
		},
		// higher term, same logs -> true
		{
			RequestVoteArgs{
				Term:         6,
				CandidateId:  1235,
				LastLogIndex: 2,
				LastLogTerm:  2,
			},
			RequestVoteReply{
				Term:        6,
				VoteGranted: true,
			},
		},
		// higher LastLogTerm -> false (already voted)
		{
			RequestVoteArgs{
				Term:         5,
				CandidateId:  1235,
				LastLogIndex: 10,
				LastLogTerm:  4,
			},
			RequestVoteReply{
				Term:        5,
				VoteGranted: false,
			},
		},
	}

	for _, tt := range tests {
		n := SetupNode(1234)

		n.Log = append(n.Log, LogEntry{Term: 1, Command: "1"})
		n.Log = append(n.Log, LogEntry{Term: 2, Command: "2"})
		n.CurrentTerm = 5
		n.VotedFor = 1234
		n.State = CANDIDATE

		actualReply := RequestVoteReply{}
		err := n.RequestVote(&tt.args, &actualReply)
		if err != nil {
			t.Errorf("Error when requesting vote: %v\n", err)
		}

		if tt.reply.Term != actualReply.Term {
			t.Errorf("Mismatching Term: expected %v, got %v\n", tt.reply.Term, actualReply.Term)
		}
		if tt.reply.VoteGranted != actualReply.VoteGranted {
			t.Errorf("Mismatching Vote: expected %v, got %v\n", tt.reply.VoteGranted, actualReply.VoteGranted)
		}
	}
}
