package raft

import (
	"testing"
)

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

