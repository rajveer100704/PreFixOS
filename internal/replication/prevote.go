package replication

import (
	"time"
)

// PreVoteArgs represents PreVote RPC request parameters.
type PreVoteArgs struct {
	Term         uint64
	CandidateID  int64
	LastLogIndex uint64
	LastLogTerm  uint64
}

// PreVoteReply represents PreVote RPC response parameters.
type PreVoteReply struct {
	Term        uint64
	VoteGranted bool
}

// PreVoteManager prevents re-joining partitioned nodes from causing disruptive elections.
type PreVoteManager struct{}

// NewPreVoteManager initializes a new PreVote manager instance.
func NewPreVoteManager() *PreVoteManager {
	return &PreVoteManager{}
}

// HandlePreVote processes an incoming PreVote request without incrementing currentTerm.
func (rn *RaftNode) HandlePreVote(args *PreVoteArgs, reply *PreVoteReply) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	reply.Term = rn.currentTerm
	reply.VoteGranted = false

	// Rule 1: Reject if candidate's term < currentTerm
	if args.Term < rn.currentTerm {
		return
	}

	// Rule 2: Check if candidate's log is at least as up-to-date as receiver's log
	lastLogIndex := uint64(len(rn.log) - 1)
	lastLogTerm := rn.log[lastLogIndex].Term

	logUpToDate := false
	if args.LastLogTerm > lastLogTerm {
		logUpToDate = true
	} else if args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex {
		logUpToDate = true
	}

	// Rule 3: Grant PreVote if lease has expired or heartbeat timed out
	if time.Since(rn.lastHeartbeat) > rn.electionTimeout && logUpToDate {
		reply.VoteGranted = true
	}
}

// TimeoutNowArgs represents TimeoutNow RPC parameters for graceful leader transfer.
type TimeoutNowArgs struct {
	Term     uint64
	LeaderID int64
}

// TimeoutNowReply represents TimeoutNow RPC reply parameters.
type TimeoutNowReply struct {
	Success bool
}

// HandleTimeoutNow immediately triggers an election for graceful leader transfer.
func (rn *RaftNode) HandleTimeoutNow(args *TimeoutNowArgs, reply *TimeoutNowReply) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term < rn.currentTerm {
		reply.Success = false
		return
	}

	// Reset election timer and transition to Candidate immediately
	rn.role = Candidate
	rn.currentTerm++
	rn.votedFor = rn.id
	_ = rn.savePersistentState()
	reply.Success = true
}
