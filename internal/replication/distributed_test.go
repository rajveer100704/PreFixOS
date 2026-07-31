package replication

import (
	"testing"
	"time"

	"prefixos/internal/memory"
	"prefixos/internal/radix"
)

func TestDistributed_AppendEntriesAndLogMatching(t *testing.T) {
	dataDir := t.TempDir()

	follower, err := NewRaftNode(1, []int64{1, 2}, dataDir)
	if err != nil {
		t.Fatalf("failed creating follower node: %v", err)
	}
	defer follower.Close()

	// Append entry 1
	cmd1 := &InsertCommand{Tokens: []int32{10, 20, 30}}
	payload1, _ := cmd1.Marshal()

	args := &AppendEntriesArgs{
		Term:         1,
		LeaderID:     2,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      []LogEntry{{Index: 1, Term: 1, Payload: payload1}},
		LeaderCommit: 1,
	}
	reply := &AppendEntriesReply{}
	follower.AppendEntries(args, reply)

	if !reply.Success {
		t.Fatalf("expected AppendEntries success")
	}
	if len(follower.log) != 2 {
		t.Fatalf("expected log length 2, got %d", len(follower.log))
	}
}

func TestDistributed_AsyncApplyQueue(t *testing.T) {
	bm := memory.NewBlockManager(100)
	tree := radix.NewTree(bm)
	sm := &StateMachine{Tree: tree, BM: bm}

	aq := NewApplyQueue(sm, 10)
	defer aq.Close()

	cmd := &InsertCommand{Tokens: []int32{100, 200, 300}}
	payload, _ := cmd.Marshal()

	entry := LogEntry{Index: 1, Term: 1, Payload: payload}
	if err := aq.PushCommittedEntry(entry); err != nil {
		t.Fatalf("failed pushing committed entry: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if aq.GetAppliedIndex() != 1 {
		t.Fatalf("expected applied index 1, got %d", aq.GetAppliedIndex())
	}

	matchedLen, _ := tree.FindLongestPrefix([]int32{100, 200, 300})
	if matchedLen != 3 {
		t.Fatalf("expected matched prefix length 3 in state machine tree, got %d", matchedLen)
	}
}

func TestDistributed_JointConsensusAndLearnerPromotion(t *testing.T) {
	cs := NewConfigState([]int64{1, 2, 3})

	// Add node 4 as Learner
	cs.AddLearner(4)
	if cs.OldMembers[4].Role != Learner {
		t.Fatalf("expected node 4 to be Learner")
	}

	// Promote Learner node 4 to Voter
	if err := cs.PromoteLearner(4); err != nil {
		t.Fatalf("failed promoting Learner: %v", err)
	}
	if cs.OldMembers[4].Role != Voter {
		t.Fatalf("expected node 4 to be Voter after promotion")
	}

	// Joint Consensus transition
	cs.EnterJointConsensus([]int64{1, 2, 3, 4})
	if !cs.InJointConfig {
		t.Fatalf("expected in joint config state")
	}

	cs.FinalizeJointConsensus()
	if cs.InJointConfig {
		t.Fatalf("expected joint config finalized")
	}
}

func TestDistributed_ChunkedSnapshotStreaming(t *testing.T) {
	dataDir := t.TempDir()
	streamer := NewSnapshotStreamer(dataDir)

	chunk1 := &InstallSnapshotArgs{
		Term:              1,
		LeaderID:          2,
		LastIncludedIndex: 100,
		LastIncludedTerm:  1,
		Offset:            0,
		Data:              []byte("chunk_part_1_"),
		Done:              false,
	}
	if err := streamer.ReceiveChunk(chunk1); err != nil {
		t.Fatalf("failed receiving chunk 1: %v", err)
	}

	chunk2 := &InstallSnapshotArgs{
		Term:              1,
		LeaderID:          2,
		LastIncludedIndex: 100,
		LastIncludedTerm:  1,
		Offset:            13,
		Data:              []byte("chunk_part_2"),
		Done:              true,
	}
	if err := streamer.ReceiveChunk(chunk2); err != nil {
		t.Fatalf("failed receiving final chunk: %v", err)
	}
}

func TestDistributed_RepeatedLeaderElectionCrashLoop(t *testing.T) {
	dataDir := t.TempDir()

	// Simulate 50 crash & election iterations
	for i := 0; i < 50; i++ {
		node, err := NewRaftNode(1, []int64{1, 2, 3}, dataDir)
		if err != nil {
			t.Fatalf("iteration %d: failed creating node: %v", i, err)
		}
		node.mu.Lock()
		node.currentTerm++
		node.votedFor = 1
		_ = node.savePersistentState()
		node.mu.Unlock()

		node.Close()
	}
}
