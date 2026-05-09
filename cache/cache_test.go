package cache_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apstndb/go-googlesql-executequery/cache"
)

func TestDefaultUnderUserCacheDir(t *testing.T) {
	// Redirect both XDG_CACHE_HOME (Linux) and HOME (macOS — which
	// reads ~/Library/Caches via $HOME) so this test works on both
	// platforms without depending on os.UserCacheDir's per-OS rules.
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("HOME", dir)

	got, err := cache.Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	if !strings.HasPrefix(got, dir+string(filepath.Separator)) {
		t.Errorf("Default() = %q; want a path under %q", got, dir)
	}
	if !strings.Contains(got, filepath.Join("go-googlesql-executequery", "wazero")) {
		t.Errorf("Default() = %q; want path to include go-googlesql-executequery/wazero", got)
	}
}

func TestResolveExplicit(t *testing.T) {
	got, err := cache.Resolve("/explicit/path")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "/explicit/path" {
		t.Errorf("Resolve(%q) = %q; want it back unchanged", "/explicit/path", got)
	}
}

func TestEnsureSafeRejectsNonDirectory(t *testing.T) {
	dir := t.TempDir()
	conflict := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(conflict, []byte("hi"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := cache.EnsureSafe(conflict); err == nil {
		t.Fatalf("expected error when target is a file")
	}
}

func TestEnsureSafeCreatesMissingDir(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", root)
	target := filepath.Join(root, "go-googlesql-executequery", "wazero", "vtest")
	got, err := cache.EnsureSafe(target)
	if err != nil {
		t.Fatalf("EnsureSafe: %v", err)
	}
	if got == "" {
		t.Fatal("EnsureSafe returned empty path")
	}
	info, err := os.Stat(got)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected directory, got %v", info.Mode())
	}
}

func TestEnsureSafeRejectsSymlinkOutOfRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", root)
	outside := t.TempDir() // separate tempdir; outside the cache root
	link := filepath.Join(root, "go-googlesql-executequery", "wazero", "vtest")
	if err := os.MkdirAll(filepath.Dir(link), 0o700); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if _, err := cache.EnsureSafe(link); err == nil {
		t.Fatalf("expected EnsureSafe to refuse symlink-out-of-root")
	}
}
