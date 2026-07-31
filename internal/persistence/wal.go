package persistence

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sync"

	"prefixos/internal/interfaces"
)

var (
	ErrCorruptWALRecord = errors.New("corrupt WAL record: checksum mismatch")
	ErrInvalidWALPath   = errors.New("invalid WAL file path")
)

const (
	OpInsert byte = 1
	OpEvict  byte = 2
	OpDelete byte = 3
)

// WALManager manages high-performance append-only Write-Ahead Logging and replay.
type WALManager struct {
	mu            sync.Mutex
	file          *os.File
	path          string
	sequence      uint64
	syncImmediate bool
}

// NewWALManager creates or opens an existing WAL file for append-only mutation logging.
func NewWALManager(path string, syncImmediate bool) (*WALManager, error) {
	if path == "" {
		return nil, ErrInvalidWALPath
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file: %w", err)
	}

	return &WALManager{
		file:          file,
		path:          path,
		syncImmediate: syncImmediate,
	}, nil
}

// EncodeRecord serializes a WALEntry into binary representation with CRC32 checksum.
// Layout: [Sequence: 8B][Type: 1B][Length: 4B][Payload: NB][Checksum: 4B]
func EncodeRecord(entry interfaces.WALEntry) []byte {
	payloadLen := len(entry.Payload)
	buf := make([]byte, 8+1+4+payloadLen+4)

	binary.BigEndian.PutUint64(buf[0:8], entry.Sequence)
	buf[8] = entry.Type
	binary.BigEndian.PutUint32(buf[9:13], uint32(payloadLen))
	copy(buf[13:13+payloadLen], entry.Payload)

	// Compute CRC32 over header + payload
	checksum := crc32.ChecksumIEEE(buf[:13+payloadLen])
	binary.BigEndian.PutUint32(buf[13+payloadLen:], checksum)

	return buf
}

// DecodeRecord reads and verifies a single binary WAL record from an io.Reader.
func DecodeRecord(r io.Reader) (interfaces.WALEntry, error) {
	header := make([]byte, 13)
	if _, err := io.ReadFull(r, header); err != nil {
		return interfaces.WALEntry{}, err
	}

	seq := binary.BigEndian.Uint64(header[0:8])
	opType := header[8]
	payloadLen := binary.BigEndian.Uint32(header[9:13])

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return interfaces.WALEntry{}, fmt.Errorf("failed reading WAL payload: %w", err)
	}

	var checksumBuf [4]byte
	if _, err := io.ReadFull(r, checksumBuf[:]); err != nil {
		return interfaces.WALEntry{}, fmt.Errorf("failed reading WAL checksum: %w", err)
	}

	expectedChecksum := binary.BigEndian.Uint32(checksumBuf[:])

	// Calculate checksum over header + payload
	checksumData := append(header, payload...)
	calculatedChecksum := crc32.ChecksumIEEE(checksumData)

	if calculatedChecksum != expectedChecksum {
		return interfaces.WALEntry{}, ErrCorruptWALRecord
	}

	return interfaces.WALEntry{
		Sequence: seq,
		Type:     opType,
		Payload:  payload,
		Checksum: expectedChecksum,
	}, nil
}

// AppendWAL writes a mutation record to the WAL file.
func (w *WALManager) AppendWAL(entry interfaces.WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.sequence++
	entry.Sequence = w.sequence

	encoded := EncodeRecord(entry)
	if _, err := w.file.Write(encoded); err != nil {
		return fmt.Errorf("failed writing WAL entry: %w", err)
	}

	if w.syncImmediate {
		return w.file.Sync()
	}

	return nil
}

// ReadAllRecords replays all valid records from the WAL file.
func (w *WALManager) ReadAllRecords() ([]interfaces.WALEntry, uint64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return nil, 0, fmt.Errorf("failed seeking WAL start: %w", err)
	}

	var entries []interfaces.WALEntry
	var maxSeq uint64

	for {
		entry, err := DecodeRecord(w.file)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return entries, maxSeq, err
		}
		entries = append(entries, entry)
		if entry.Sequence > maxSeq {
			maxSeq = entry.Sequence
		}
	}

	w.sequence = maxSeq
	return entries, maxSeq, nil
}

// Close flushes and closes the underlying WAL file.
func (w *WALManager) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		_ = w.file.Sync()
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}
