# consistent-hash examples

Offline usage examples (no network required).

Build first:

```bash
export GOTOOLCHAIN=local CGO_ENABLED=0
go build -o /tmp/consistent-hash .
```

Look up which node owns a key:

```bash
/tmp/consistent-hash get -key user:42 -nodes node-a,node-b,node-c
/tmp/consistent-hash get -key user:42 -nodes node-a,node-b,node-c -replicas 100
```

The command prints the owning node and exits 0. An empty node list or empty
ring prints an error and exits non-zero (it never panics).
