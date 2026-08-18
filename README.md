# consistent-hash

Persistent consistent-hash ring with virtual nodes, snapshot persistence, and
balance measurement. Supports adding/removing nodes with minimal key movement,
atomic checkpoint/recovery, and distribution quality analysis. Pure Go, standard
library only.

## Build & Test

```bash
export GOTOOLCHAIN=local CGO_ENABLED=0
go build ./...
go test ./...
```

## Usage

### In-memory lookup

```bash
consistent-hash get -key "user:42" -nodes "srv-01,srv-02,srv-03" -replicas 100
```

### Persistent mode (checkpoint/recovery)

```bash
# Session 1: add nodes and checkpoint
consistent-hash serve -dir ./data -replicas 100 -add "srv-01,srv-02,srv-03" -key "user:42" -checkpoint

# Session 2: recovered from snapshot, add a new node
consistent-hash serve -dir ./data -replicas 100 -add "srv-04" -key "user:42" -checkpoint
```

## Directory Structure

```
internal/hash       FNV-1a hashing, virtual positions, distribution spread
internal/node       Name normalization, validation, set operations
internal/ring       Consistent-hash ring (Add, Remove, Get, Members)
internal/snapshot   Binary snapshot persistence (magic, version, CRC32)
internal/store      Persistent store (Open, Checkpoint, MinimalMovement)
internal/balance    Distribution measurement, migration planning
```

## Persistence

The `serve` command persists ring membership as a binary snapshot with CRC32
integrity. On recovery, the ring is deterministically rebuilt from the member
list (virtual nodes are computed from the hash function).

## Properties

- **Consistent**: same key maps to same node across lookups
- **Minimal movement**: adding/removing a node moves only affected keys
- **Recoverable**: snapshot restores identical lookup behavior after restart
- **Balanced**: virtual nodes distribute load across physical nodes
