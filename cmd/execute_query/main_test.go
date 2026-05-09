package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary compiles the CLI into a per-test tempdir and returns
// the resulting path. Tests are skipped under -short.
func buildBinary(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("CLI golden tests build the binary; skipped under -short")
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "execute_query")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build: %v", err)
	}
	return bin
}

// runBinary invokes bin with args, returning (stdout, stderr, exit code).
func runBinary(t *testing.T, bin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	// Re-use a per-process cache dir so successive tests warm-start
	// the wasm runtime. The cache package keys by go-googlesql
	// version, so this is safe across test files.
	cacheDir := filepath.Join(os.TempDir(), "go-googlesql-executequery-cli-test-cache")
	cmd.Env = append(os.Environ(),
		"XDG_CACHE_HOME="+filepath.Join(cacheDir, "xdg"),
		"HOME="+cacheDir,
	)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			t.Fatalf("run: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), code
}

func TestCLIParseMode(t *testing.T) {
	bin := buildBinary(t)
	stdout, stderr, code := runBinary(t, bin, "--mode=parse", "SELECT 1+1")
	if code != 0 {
		t.Fatalf("exit %d, stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "QueryStatement") {
		t.Errorf("stdout missing QueryStatement: %q", stdout)
	}
}

func TestCLIDescribeTPCH(t *testing.T) {
	bin := buildBinary(t)
	stdout, _, code := runBinary(t, bin, "--catalog=tpch", "DESCRIBE Orders")
	if code != 0 {
		t.Fatalf("exit %d, stdout=%q", code, stdout)
	}
	if !strings.Contains(stdout, "O_ORDERKEY") {
		t.Errorf("stdout missing O_ORDERKEY: %q", stdout)
	}
}

func TestCLIExecuteRejected(t *testing.T) {
	bin := buildBinary(t)
	_, stderr, code := runBinary(t, bin, "--mode=execute", "SELECT 1")
	if code != 2 {
		t.Errorf("expected exit 2 for unsupported mode, got %d (stderr=%q)", code, stderr)
	}
	if !strings.Contains(stderr, "execute") {
		t.Errorf("stderr should mention execute: %q", stderr)
	}
}

func TestCLIWebRejected(t *testing.T) {
	bin := buildBinary(t)
	_, stderr, code := runBinary(t, bin, "--web", "SELECT 1")
	if code != 2 {
		t.Errorf("expected exit 2 for --web, got %d (stderr=%q)", code, stderr)
	}
	if !strings.Contains(stderr, "--web") {
		t.Errorf("stderr should mention --web: %q", stderr)
	}
}

func TestCLIStdinInput(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "--mode=unparse", "-")
	cmd.Stdin = strings.NewReader("SELECT 42")
	cacheDir := filepath.Join(os.TempDir(), "go-googlesql-executequery-cli-test-cache")
	cmd.Env = append(os.Environ(),
		"XDG_CACHE_HOME="+filepath.Join(cacheDir, "xdg"),
		"HOME="+cacheDir,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("cmd.Output: %v", err)
	}
	if !strings.Contains(string(out), "42") {
		t.Errorf("stdin SQL did not appear in output: %q", string(out))
	}
}
