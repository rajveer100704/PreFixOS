package replication

import (
	"testing"
)

func TestRaftCore_PersistentStateSavingAndLoading(t *testing.T) {
	dataDir := t.TempDir()

	// 1. Initialize node 1
	node1, err := NewRaftNode(1, []int64{1, 2, 3}, dataDir)
	if err != nil {
		t.Fatalf("failed creating RaftNode: %v", err)
	}

	node1.mu.Lock()
	node1.currentTerm = 5
	node1.votedFor = 2
	node1.log = append(node1.log, LogEntry{Index: 1, Term: 5, Payload: []byte("cmd1")})
	_ = node1.savePersistentState()
	node1.mu.Unlock()

	node1.Close()

	// 2. Restart node 1 from persistent disk state
	node1Restarted, err := NewRaftNode(1, []int64{1, 2, 3}, dataDir)
	if err != nil {
		t.Fatalf("failed restarting RaftNode: %v", err)
	}
	defer node1Restarted.Close()

	if node1Restarted.currentTerm != 5 {
		t.Fatalf("expected term 5 after restart, got %d", node1Restarted.currentTerm)
	}
	if node1Restarted.votedFor != 2 {
		t.Fatalf("expected votedFor 2 after restart, got %d", node1Restarted.votedFor)
	}
	if len(node1Restarted.log) != 2 {
		t.Fatalf("expected log length 2 after restart, got %d", len(node1Restarted.log))
	}
}

func TestRaftCore_RequestVoteSafety(t *testing.T) {
	dataDir := t.TempDir()

	node, err := NewRaftNode(1, []int64{1, 2}, dataDir)
	if err != nil {
		t.Fatalf("failed creating node: %v", err)
	}
	defer node.Close()

	node.mu.Lock()
	node.currentTerm = 2
	node.mu.Unlock()

	// Rule: Term < currentTerm -> Reject vote
	argsLow := &RequestVoteArgs{Term: 1, CandidateID: 2, LastLogIndex: 0, LastLogTerm: 0}
	replyLow := &RequestVoteReply{}
	node.RequestVote(argsLow, replyLow)

	if replyLow.VoteGranted {
		t.Fatalf("expected vote rejected for lower term")
	}

	// Rule: Valid candidate term -> Grant vote
	argsValid := &RequestVoteArgs{Term: 3, CandidateID: 2, LastLogIndex: 0, LastLogTerm: 0}
	replyValid := &RequestVoteReply{}
	node.RequestVote(argsValid, replyValid)

	if !replyValid.VoteGranted {
		t.Fatalf("expected vote granted for valid candidate")
	}

	// Rule: Second candidate in same term -> Reject vote (Election Safety)
	argsDup := &RequestVoteArgs{Term: 3, CandidateID: 3, LastLogIndex: 0, LastLogTerm: 0}
	replyDup := &RequestVoteReply{}
	node.RequestVote(argsDup, replyDup)

	if replyDup.VoteGranted {
		t.Fatalf("expected duplicate vote rejected in same term (Election Safety Violation)")
	}
}
