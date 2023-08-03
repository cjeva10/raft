package raft

import (
	"testing"


)

func TestAppendEntries(t *testing.T) {
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
