package node

import (
	"errors"
	"testing"
)

func TestNormalizeTrim(t *testing.T) {
	if got := Normalize("  Node-A "); got != "node-a" {
		t.Errorf("Normalize = %q, want %q", got, "node-a")
	}
}

func TestValidateEmpty(t *testing.T) {
	if err := Validate(""); err != ErrEmptyName {
		t.Errorf("err = %v, want ErrEmptyName", err)
	}
	if err := Validate("ok"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDedupe(t *testing.T) {
	got := NormalizeAll([]string{"A", "a", " B ", "", "A", "c"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestNormalizeAllDropsEmpty(t *testing.T) {
	if got := NormalizeAll([]string{"", "   ", ""}); len(got) != 0 {
		t.Errorf("expected no names, got %v", got)
	}
}

func TestValidateErrorIsWrapped(t *testing.T) {
	if !errors.Is(Validate(""), ErrEmptyName) {
		t.Error("Validate(\"\") should be ErrEmptyName")
	}
}

func TestValidateHashChar(t *testing.T) {
	if err := Validate("bad#name"); err != ErrInvalidChar {
		t.Errorf("err = %v, want ErrInvalidChar", err)
	}
}

func TestSort(t *testing.T) {
	names := []string{"c", "a", "b"}
	Sort(names)
	if names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Errorf("Sort result = %v", names)
	}
}

func TestContains(t *testing.T) {
	names := []string{"x", "y", "z"}
	if !Contains(names, "y") {
		t.Error("should contain y")
	}
	if Contains(names, "w") {
		t.Error("should not contain w")
	}
}

func TestDiffNodes(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"b"}
	diff := Diff(a, b)
	if len(diff) != 2 {
		t.Fatalf("Diff len = %d, want 2", len(diff))
	}
	if diff[0] != "a" || diff[1] != "c" {
		t.Errorf("Diff = %v, want [a c]", diff)
	}
}
