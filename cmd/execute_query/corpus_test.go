//go:build corpus

package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCorpus walks third_party/googlesql/googlesql/examples and runs
// the built CLI in --mode=parse for each .sql file under tpch/queries
// and pipe_queries. The submodule pin determines which queries are
// exercised, so this gives stable golden coverage tied to the
// upstream commit.
//
// Build-tagged `corpus` so the default `go test ./...` skips it (the
// corpus is large and slows down PR-time signal).
func TestCorpus(t *testing.T) {
	bin := buildBinary(t)

	root := repoRoot(t)
	dirs := []string{
		filepath.Join(root, "third_party", "googlesql", "googlesql", "examples", "tpch", "queries"),
		filepath.Join(root, "third_party", "googlesql", "googlesql", "examples", "pipe_queries"),
	}
	any := false
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Logf("skip %q: %v", dir, err)
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
				continue
			}
			any = true
			path := filepath.Join(dir, e.Name())
			t.Run(filepath.Base(dir)+"/"+e.Name(), func(t *testing.T) {
				_, stderr, code := runBinary(t, bin,
					"--mode=parse",
					"--catalog=tpch",
					"@"+path,
				)
				if code != 0 {
					t.Errorf("exit %d for %s: %s", code, path, stderr)
				}
			})
		}
	}
	if !any {
		t.Skip("no corpus files found; populate the submodule with `git submodule update --init`")
	}
}

// repoRoot returns the repository root by walking up from the test
// binary's CWD until it finds go.mod.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find go.mod above %q", dir)
		}
		dir = parent
	}
}
