package snapshot

import (
	"encoding/binary"
	"hash/crc32"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ring.snap")

	state := &RingState{
		Replicas: 150,
		Nodes:    []string{"node-a", "node-b", "node-c"},
	}

	if err := Save(path, state); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Replicas != 150 {
		t.Errorf("Replicas = %d, want 150", loaded.Replicas)
	}
	if len(loaded.Nodes) != 3 {
		t.Fatalf("Nodes len = %d, want 3", len(loaded.Nodes))
	}
	for i, n := range loaded.Nodes {
		if n != state.Nodes[i] {
			t.Errorf("Nodes[%d] = %q, want %q", i, n, state.Nodes[i])
		}
	}
}

func TestSaveEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.snap")

	state := &RingState{Replicas: 100, Nodes: nil}
	if err := Save(path, state); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Replicas != 100 || len(loaded.Nodes) != 0 {
		t.Errorf("loaded = %+v, want replicas=100, empty nodes", loaded)
	}
}

func TestLoadBadMagic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.snap")
	os.WriteFile(path, []byte("BADMxxxxxxxxxxxxxxxxxxxxx"), 0644)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestLoadCorruptCRC(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ring.snap")

	state := &RingState{Replicas: 10, Nodes: []string{"x"}}
	Save(path, state)

	data, _ := os.ReadFile(path)
	data[6] ^= 0xFF
	os.WriteFile(path, data, 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected CRC error")
	}
}

func TestLoadTruncated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ring.snap")

	state := &RingState{Replicas: 10, Nodes: []string{"long-node-name-here"}}
	Save(path, state)

	data, _ := os.ReadFile(path)
	os.WriteFile(path, data[:len(data)/2], 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected truncation error")
	}
}

func TestLoadUnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ring.snap")

	state := &RingState{Replicas: 10, Nodes: nil}
	Save(path, state)

	data, _ := os.ReadFile(path)
	data[4] = 99 // version byte
	// recompute CRC
	payload := data[:len(data)-4]
	binary.LittleEndian.PutUint32(data[len(data)-4:], crc32.ChecksumIEEE(payload))
	os.WriteFile(path, data, 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected unsupported version error")
	}
}

func TestSaveAtomicity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "atomic.snap")

	state := &RingState{Replicas: 5, Nodes: []string{"a"}}
	Save(path, state)

	tmp := path + ".tmp"
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Error(".tmp should not remain")
	}
}
