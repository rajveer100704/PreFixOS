package interfaces

// ReplicationEngine defines Raft consensus and node clustering contracts
type ReplicationEngine interface {
	IsLeader() bool
	GetLeaderAddress() string
	ReplicateLog(entry WALEntry) error
}
