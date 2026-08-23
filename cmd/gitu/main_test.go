package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLooksSensitive(t *testing.T) {
	cases := map[string]bool{"notes/plan.md": false, ".env": true, ".env.local": true, "certs/site.pem": true, "docs/token-rotation.md": true, ".aws/config": true}
	for path, want := range cases {
		if got := looksSensitive(path); got != want {
			t.Errorf("looksSensitive(%q) = %v, want %v", path, got, want)
		}
	}
}
func TestComparePaths(t *testing.T) {
	dir := t.TempDir()
	left, right := filepath.Join(dir, "left"), filepath.Join(dir, "right")
	if err := os.WriteFile(left, []byte("one"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(right, []byte("one"), 0600); err != nil {
		t.Fatal(err)
	}
	state, err := comparePaths(left, right)
	if err != nil || state != "up to date" {
		t.Fatalf("got (%q, %v)", state, err)
	}
	if err := os.WriteFile(right, []byte("two"), 0600); err != nil {
		t.Fatal(err)
	}
	state, err = comparePaths(left, right)
	if err != nil || state != "changed" {
		t.Fatalf("got (%q, %v)", state, err)
	}
}
func TestSlug(t *testing.T) {
	if got := slug("Client / Project Notes!"); got != "client-project-notes" {
		t.Fatalf("got %q", got)
	}
}

func TestBranchKeySeparatesSanitizedNames(t *testing.T) {
	if branchKey("feature/docs") == branchKey("feature docs") {
		t.Fatal("branch key must include a stable discriminator")
	}
}
