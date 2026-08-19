// Package snapshot persists and loads the state of a consistent-hash ring.
//
// File format (little-endian):
//
//	[magic 4B]["CHRS"] [version 1B][0x01] [replicas uint32]
//	[node_count uint32]
//	for each node:
//	    [name_len uint16] [name ...]
//	[crc32 4B] (IEEE over everything before this field)
//
// The snapshot captures the full membership list. On load, the ring is rebuilt
// from the member list (virtual nodes are deterministic given the hash
// function and replicas count).
package snapshot

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
)

var fileMagic = [4]byte{'C', 'H', 'R', 'S'}

const currentVersion = 1

var (
	ErrBadMagic           = errors.New("snapshot: invalid magic bytes")
	ErrUnsupportedVersion = errors.New("snapshot: unsupported version")
	ErrCorrupt            = errors.New("snapshot: data corrupted (CRC mismatch)")
	ErrTruncated          = errors.New("snapshot: file truncated")
)

// RingState holds the serializable state of a consistent-hash ring.
type RingState struct {
	Replicas int
	Nodes    []string
}

// Save writes the ring state to path atomically (write tmp + rename).
func Save(path string, state *RingState) error {
	data, err := encode(state)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("snapshot: create tmp: %w", err)
	}
	defer func() {
		f.Close()
		os.Remove(tmp)
	}()
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("snapshot: write: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("snapshot: sync: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("snapshot: close: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("snapshot: rename: %w", err)
	}
	return nil
}

// Load reads a ring state snapshot from path.
func Load(path string) (*RingState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: read: %w", err)
	}
	return decode(data)
}

func encode(state *RingState) ([]byte, error) {
	buf := make([]byte, 0, 256)

	// header
	buf = append(buf, fileMagic[:]...)
	buf = append(buf, currentVersion)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(state.Replicas))

	// nodes
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(state.Nodes)))
	for _, n := range state.Nodes {
		if len(n) > 65535 {
			return nil, fmt.Errorf("snapshot: node name too long: %d bytes", len(n))
		}
		buf = binary.LittleEndian.AppendUint16(buf, uint16(len(n)))
		buf = append(buf, n...)
	}

	// CRC
	crc := crc32.ChecksumIEEE(buf)
	buf = binary.LittleEndian.AppendUint32(buf, crc)
	return buf, nil
}

func decode(data []byte) (*RingState, error) {
	// minimum: magic(4) + version(1) + replicas(4) + count(4) + crc(4) = 17
	if len(data) < 17 {
		return nil, ErrTruncated
	}

	// verify CRC
	payload := data[:len(data)-4]
	storedCRC := binary.LittleEndian.Uint32(data[len(data)-4:])
	if crc32.ChecksumIEEE(payload) != storedCRC {
		return nil, ErrCorrupt
	}

	pos := 0

	// magic
	var m [4]byte
	copy(m[:], data[pos:pos+4])
	pos += 4
	if m != fileMagic {
		return nil, ErrBadMagic
	}

	// version
	ver := data[pos]
	pos++
	if ver != currentVersion {
		return nil, fmt.Errorf("%w: got %d", ErrUnsupportedVersion, ver)
	}

	// replicas
	replicas := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
	pos += 4

	// node count
	count := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
	pos += 4

	nodes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if pos+2 > len(payload) {
			return nil, ErrTruncated
		}
		nameLen := int(binary.LittleEndian.Uint16(data[pos : pos+2]))
		pos += 2
		if pos+nameLen > len(payload) {
			return nil, ErrTruncated
		}
		nodes = append(nodes, string(data[pos:pos+nameLen]))
		pos += nameLen
	}

	return &RingState{
		Replicas: replicas,
		Nodes:    nodes,
	}, nil
}
