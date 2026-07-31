package replication

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// NodeRole represents the current state of a Raft node.
type NodeRole int

const (
	Follower NodeRole = iota
	Candidate
	Leader
)

func (r NodeRole) String() string {
	switch r {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

// PersistentState stores the mandatory persistent Raft state variables across restarts.
type PersistentState struct {
	CurrentTerm uint64     `json:"current_term"`
	VotedFor    int64      `json:"voted_for"`
	Log         []LogEntry `json:"log"`
}

// RaftNode implements the core Raft consensus protocol state machine.
type RaftNode struct {
	mu sync.Mutex

	// Persistent state on all nodes (Updated before responding to RPCs)
	id          int64
	currentTerm uint64
	votedFor    int64
	log         []LogEntry
	dataDir     string

	// Volatile state on all nodes
	commitIndex uint64
	lastApplied uint64
	role        NodeRole

	// Volatile state on leaders (Reinitialized after election)
	nextIndex  map[int64]uint64
	matchIndex map[int64]uint64

	// Cluster configuration
	peers []int64

	// Timers and channels
	electionTimeout time.Duration
	lastHeartbeat   time.Time
	stopChan        chan struct{}
}

// NewRaftNode initializes a new Raft consensus node with persistent state storage.
func NewRaftNode(id int64, peers []int64, dataDir string) (*RaftNode, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating Raft data directory: %w", err)
	}

	node := &RaftNode{
		id:              id,
		peers:           peers,
		dataDir:         dataDir,
		votedFor:        -1,
		role:            Follower,
		nextIndex:       make(map[int64]uint64),
		matchIndex:      make(map[int64]uint64),
		stopChan:        make(chan struct{}),
		electionTimeout: time.Duration(150+rand.Intn(150)) * time.Millisecond,
		lastHeartbeat:   time.Now(),
	}

	// 1. Load persistent state if present
	if err := node.loadPersistentState(); err != nil {
		return nil, fmt.Errorf("failed loading persistent Raft state: %w", err)
	}

	return node, nil
}

// persistentFilePath returns the absolute path to the raft_state.json file.
func (rn *RaftNode) persistentFilePath() string {
	return filepath.Join(rn.dataDir, fmt.Sprintf("raft_state_%d.json", rn.id))
}

// savePersistentState atomically saves mandatory persistent Raft state (currentTerm, votedFor, log) to disk.
func (rn *RaftNode) savePersistentState() error {
	state := PersistentState{
		CurrentTerm: rn.currentTerm,
		VotedFor:    rn.votedFor,
		Log:         rn.log,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	tmpFile := rn.persistentFilePath() + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, rn.persistentFilePath())
}

// loadPersistentState restores mandatory persistent Raft state from disk.
func (rn *RaftNode) loadPersistentState() error {
	path := rn.persistentFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			rn.currentTerm = 0
			rn.votedFor = -1
			rn.log = []LogEntry{{Index: 0, Term: 0}} // Dummy sentinel entry
			return nil
		}
		return err
	}

	var state PersistentState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	rn.currentTerm = state.CurrentTerm
	rn.votedFor = state.VotedFor
	rn.log = state.Log
	if len(rn.log) == 0 {
		rn.log = []LogEntry{{Index: 0, Term: 0}}
	}
	return nil
}

// GetState returns the current term and role of the Raft node.
func (rn *RaftNode) GetState() (uint64, bool) {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.currentTerm, rn.role == Leader
}

// RequestVoteArgs represents RequestVote RPC request parameters.
type RequestVoteArgs struct {
	Term         uint64
	CandidateID  int64
	LastLogIndex uint64
	LastLogTerm  uint64
}

// RequestVoteReply represents RequestVote RPC reply parameters.
type RequestVoteReply struct {
	Term        uint64
	VoteGranted bool
}

// RequestVote handles incoming candidate vote requests.
func (rn *RaftNode) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	// 1. Rule: Term < currentTerm -> Reject
	if args.Term < rn.currentTerm {
		reply.Term = rn.currentTerm
		reply.VoteGranted = false
		return
	}

	// 2. Rule: If Term > currentTerm, update term and step down to Follower
	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.role = Follower
		rn.votedFor = -1
		_ = rn.savePersistentState()
	}

	reply.Term = rn.currentTerm
	reply.VoteGranted = false

	// 3. Log up-to-date check
	lastLogIndex := uint64(len(rn.log) - 1)
	lastLogTerm := rn.log[lastLogIndex].Term

	logUpToDate := false
	if args.LastLogTerm > lastLogTerm {
		logUpToDate = true
	} else if args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex {
		logUpToDate = true
	}

	// 4. Grant vote if votedFor is null or candidateID, and candidate's log is at least as up-to-date
	if (rn.votedFor == -1 || rn.votedFor == args.CandidateID) && logUpToDate {
		rn.votedFor = args.CandidateID
		reply.VoteGranted = true
		rn.lastHeartbeat = time.Now()
		_ = rn.savePersistentState()
	}
}

// AppendEntriesArgs represents AppendEntries RPC request parameters.
type AppendEntriesArgs struct {
	Term         uint64
	LeaderID     int64
	PrevLogIndex uint64
	PrevLogTerm  uint64
	Entries      []LogEntry
	LeaderCommit uint64
}

// AppendEntriesReply represents AppendEntries RPC reply parameters.
type AppendEntriesReply struct {
	Term    uint64
	Success bool
}

// AppendEntries handles log replication and heartbeat RPCs from the Leader.
func (rn *RaftNode) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	reply.Success = false
	reply.Term = rn.currentTerm

	// 1. Reply false if term < currentTerm
	if args.Term < rn.currentTerm {
		return
	}

	// 2. Update term if greater and reset to Follower
	if args.Term > rn.currentTerm || rn.role == Candidate {
		rn.currentTerm = args.Term
		rn.role = Follower
		rn.votedFor = -1
		_ = rn.savePersistentState()
	}

	rn.lastHeartbeat = time.Now()

	// 3. Reply false if log doesn't contain entry at prevLogIndex matching prevLogTerm
	if args.PrevLogIndex >= uint64(len(rn.log)) || rn.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		return
	}

	// 4. Append new entries not already in the log
	for i, entry := range args.Entries {
		idx := args.PrevLogIndex + 1 + uint64(i)
		if idx < uint64(len(rn.log)) {
			if rn.log[idx].Term != entry.Term {
				rn.log = rn.log[:idx] // Truncate conflicting entries
				rn.log = append(rn.log, entry)
			}
		} else {
			rn.log = append(rn.log, entry)
		}
	}
	_ = rn.savePersistentState()

	// 5. If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if args.LeaderCommit > rn.commitIndex {
		lastNewIndex := args.PrevLogIndex + uint64(len(args.Entries))
		if args.LeaderCommit < lastNewIndex {
			rn.commitIndex = args.LeaderCommit
		} else {
			rn.commitIndex = lastNewIndex
		}
	}

	reply.Success = true
}

// Close terminates the Raft node loops cleanly.
func (rn *RaftNode) Close() error {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	select {
	case <-rn.stopChan:
	default:
		close(rn.stopChan)
	}
	return nil
}

// GetLastApplied returns the highest log index applied to the state machine.
func (rn *RaftNode) GetLastApplied() uint64 {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	return rn.lastApplied
}
