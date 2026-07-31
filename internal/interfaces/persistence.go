package interfaces

// WALEntry represents a single mutation record in the Write-Ahead Log
type WALEntry struct {
	Sequence  uint64
	Type      byte
	Payload   []byte
	Checksum  uint32
}

// PersistenceEngine defines durability and recovery contracts
type PersistenceEngine interface {
	AppendWAL(entry WALEntry) error
	CreateSnapshot() (snapshotID string, err error)
	Recover() (lastSequence uint64, err error)
}
