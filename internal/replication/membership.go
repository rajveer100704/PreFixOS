package replication

import (
	"fmt"
	"sync"
)

// MemberRole defines voting vs non-voting cluster member roles.
type MemberRole int

const (
	Voter MemberRole = iota
	Learner
)

// ClusterMember represents a single node in the cluster configuration.
type ClusterMember struct {
	ID   int64      `json:"id"`
	Role MemberRole `json:"role"`
}

// ConfigState tracks Joint Consensus configuration transitions (C_old -> C_old,new -> C_new).
type ConfigState struct {
	mu            sync.Mutex
	OldMembers    map[int64]ClusterMember `json:"old_members"`
	NewMembers    map[int64]ClusterMember `json:"new_members"`
	InJointConfig bool                    `json:"in_joint_config"`
}

// NewConfigState initializes the cluster membership state.
func NewConfigState(initialVoters []int64) *ConfigState {
	oldMap := make(map[int64]ClusterMember)
	for _, id := range initialVoters {
		oldMap[id] = ClusterMember{ID: id, Role: Voter}
	}
	return &ConfigState{
		OldMembers:    oldMap,
		NewMembers:    make(map[int64]ClusterMember),
		InJointConfig: false,
	}
}

// AddLearner registers a non-voting Learner node to catch up on log entries.
func (cs *ConfigState) AddLearner(id int64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.OldMembers[id] = ClusterMember{ID: id, Role: Learner}
}

// PromoteLearner promotes a fully caught-up Learner node to a voting cluster member.
func (cs *ConfigState) PromoteLearner(id int64) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	member, exists := cs.OldMembers[id]
	if !exists {
		return fmt.Errorf("member %d does not exist in cluster configuration", id)
	}
	if member.Role == Voter {
		return nil // Already a voter
	}

	cs.OldMembers[id] = ClusterMember{ID: id, Role: Voter}
	return nil
}

// EnterJointConsensus initiates a Joint Consensus transition (C_old,new).
func (cs *ConfigState) EnterJointConsensus(newVoters []int64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.NewMembers = make(map[int64]ClusterMember)
	for _, id := range newVoters {
		cs.NewMembers[id] = ClusterMember{ID: id, Role: Voter}
	}
	cs.InJointConfig = true
}

// FinalizeJointConsensus completes the transition to C_new configuration.
func (cs *ConfigState) FinalizeJointConsensus() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.OldMembers = cs.NewMembers
	cs.NewMembers = make(map[int64]ClusterMember)
	cs.InJointConfig = false
}
