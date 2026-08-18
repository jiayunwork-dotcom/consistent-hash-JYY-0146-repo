package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"consistent-hash/internal/node"
	"consistent-hash/internal/ring"
	"consistent-hash/internal/store"
)

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "consistent-hash: "+format+"\n", args...)
	os.Exit(1)
}

func reorderFlags(args []string) []string {
	var flags, pos []string
	i := 0
	for i < len(args) {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && !strings.Contains(a, "=") {
				flags = append(flags, a, args[i+1])
				i += 2
			} else {
				flags = append(flags, a)
				i++
			}
		} else {
			pos = append(pos, a)
			i++
		}
	}
	return append(flags, pos...)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := reorderFlags(os.Args[2:])
	switch cmd {
	case "get":
		runGet(args)
	case "serve":
		runServe(args)
	case "help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: consistent-hash <command> [flags]

commands:
  get    in-memory lookup (no persistence)
  serve  persistent ring with checkpoint/recovery

flags (get):
  -key       lookup key
  -nodes     comma-separated node list
  -replicas  virtual nodes per physical node (default 100)

flags (serve):
  -dir        data directory for ring snapshot
  -replicas   virtual nodes per physical node (default 100)
  -add        comma-separated nodes to add
  -remove     comma-separated nodes to remove
  -key        lookup key (optional)
  -checkpoint write snapshot after operations`)
}

func runGet(args []string) {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	key := fs.String("key", "", "lookup key")
	nodesStr := fs.String("nodes", "", "comma-separated node list")
	replicas := fs.Int("replicas", 100, "virtual nodes per physical node")
	if err := fs.Parse(args); err != nil {
		fatal("%v", err)
	}
	if *key == "" {
		fatal("get requires -key")
	}
	if *nodesStr == "" {
		fatal("get requires -nodes")
	}
	nodes := node.NormalizeAll(strings.Split(*nodesStr, ","))
	if len(nodes) == 0 {
		fatal("node list is empty after normalization")
	}
	r := ring.New(*replicas)
	for _, n := range nodes {
		if err := r.Add(n); err != nil {
			fatal("add node %q: %v", n, err)
		}
	}
	owner, err := r.Get(*key)
	if err != nil {
		fatal("get: %v", err)
	}
	fmt.Println(owner)
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	dir := fs.String("dir", "", "data directory")
	replicas := fs.Int("replicas", 100, "virtual nodes per physical node")
	addStr := fs.String("add", "", "comma-separated nodes to add")
	removeStr := fs.String("remove", "", "comma-separated nodes to remove")
	key := fs.String("key", "", "lookup key (optional)")
	checkpoint := fs.Bool("checkpoint", false, "write snapshot")
	if err := fs.Parse(args); err != nil {
		fatal("%v", err)
	}
	if *dir == "" {
		fatal("serve requires -dir")
	}

	s, err := store.Open(*dir, *replicas)
	if err != nil {
		fatal("open: %v", err)
	}
	defer s.Close()

	if s.WasRecovered {
		fmt.Fprintf(os.Stderr, "info: ring recovered (%d nodes)\n", s.Len())
	}

	// add nodes
	if *addStr != "" {
		for _, n := range node.NormalizeAll(strings.Split(*addStr, ",")) {
			if err := s.Add(n); err != nil {
				fatal("add %q: %v", n, err)
			}
		}
	}

	// remove nodes
	if *removeStr != "" {
		for _, n := range node.NormalizeAll(strings.Split(*removeStr, ",")) {
			if err := s.Remove(n); err != nil {
				fatal("remove %q: %v", n, err)
			}
		}
	}

	// lookup
	if *key != "" {
		owner, err := s.Get(*key)
		if err != nil {
			fatal("get: %v", err)
		}
		fmt.Printf("key=%s owner=%s\n", *key, owner)
	}

	// checkpoint
	if *checkpoint {
		if err := s.Checkpoint(); err != nil {
			fatal("checkpoint: %v", err)
		}
		fmt.Println("checkpoint: ok")
	}

	fmt.Printf("members: %s\n", strings.Join(s.Members(), ","))
}
