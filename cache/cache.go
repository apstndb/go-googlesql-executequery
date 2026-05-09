// Package cache resolves a safe, OS-appropriate wazero
// compilation-cache directory for goccy/go-googlesql.
//
// goccy/go-googlesql ships a ~13 MB wasm module that costs roughly
// 3 s to compile to native code in CompilationModeCompiler. With a
// warm cache pointed at a stable on-disk directory, subsequent
// processes drop their Init() time to ~0.6 s. This package picks
// that directory using os.UserCacheDir conventions, isolates the
// cache per linked goccy/go-googlesql module version (so a
// dependency upgrade doesn't try to load stale precompiled wasm),
// and refuses to write anywhere that could indicate a TOCTOU attack.
package cache

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	googlesql "github.com/goccy/go-googlesql"
)

// rootDirName is the per-user cache subdirectory for this project.
const rootDirName = "go-googlesql-executequery"

// dependencyModulePath is the import path we extract a version
// from to key the cache subdir.
const dependencyModulePath = "github.com/goccy/go-googlesql"

// Default returns the canonical wazero cache directory for the
// current user. The path is keyed by the linked
// goccy/go-googlesql module version so an upgrade gets a fresh
// cache.
//
//	macOS:   ~/Library/Caches/go-googlesql-executequery/wazero/<ver>
//	Linux:   $XDG_CACHE_HOME/go-googlesql-executequery/wazero/<ver>
//	         (~/.cache/... when XDG_CACHE_HOME is unset)
//	Windows: %LocalAppData%/go-googlesql-executequery/wazero/<ver>
func Default() (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache dir: %w", err)
	}
	return filepath.Join(root, rootDirName, "wazero", linkedVersion()), nil
}

// Resolve returns dir if non-empty, otherwise Default().
func Resolve(dir string) (string, error) {
	if dir != "" {
		return dir, nil
	}
	return Default()
}

// EnsureSafe creates dir (mode 0700) if missing, and refuses to use
// it if it already exists but is not a regular directory or is a
// symlink whose target lies outside the user cache root.
//
// Returns the canonical (symlink-resolved) directory on success.
func EnsureSafe(dir string) (string, error) {
	if dir == "" {
		return "", errors.New("empty cache directory")
	}
	if info, err := os.Lstat(dir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := filepath.EvalSymlinks(dir)
			if err != nil {
				return "", fmt.Errorf("resolve symlink %q: %w", dir, err)
			}
			root, err := os.UserCacheDir()
			if err != nil {
				return "", fmt.Errorf("resolve user cache dir: %w", err)
			}
			absRoot, err := filepath.Abs(root)
			if err != nil {
				return "", fmt.Errorf("absolute path of cache root: %w", err)
			}
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return "", fmt.Errorf("absolute path of cache target: %w", err)
			}
			if !isUnder(absTarget, absRoot) {
				return "", fmt.Errorf("cache directory %q is a symlink to %q, which lies outside %q (refusing to follow)", dir, absTarget, absRoot)
			}
			dir = target
		} else if !info.IsDir() {
			return "", fmt.Errorf("cache path %q exists but is not a directory", dir)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat %q: %w", dir, err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create cache dir %q: %w", dir, err)
	}
	return dir, nil
}

// Option configures Setup.
type Option func(*setupOptions)

type setupOptions struct {
	dir         string
	disable     bool
	compileMode googlesql.CompilationMode
}

// WithDir overrides the cache directory. An empty string means
// "use Default()".
func WithDir(dir string) Option {
	return func(o *setupOptions) { o.dir = dir }
}

// Disable suppresses the on-disk cache. googlesql.Init is still
// called, in CompilationModeInterpreter (no native compile happens,
// so no cache is needed).
func Disable() Option {
	return func(o *setupOptions) { o.disable = true }
}

// WithCompilationMode lets callers override the compilation mode.
// Default is CompilationModeCompiler.
func WithCompilationMode(mode googlesql.CompilationMode) Option {
	return func(o *setupOptions) { o.compileMode = mode }
}

// Setup resolves a cache directory, creates it safely, and calls
// googlesql.Init with WithCompilationCache + WithCompilationMode.
//
// googlesql.Init is sync.Once-guarded so calling Setup more than
// once is a no-op.
func Setup(opts ...Option) error {
	o := setupOptions{compileMode: googlesql.CompilationModeCompiler}
	for _, opt := range opts {
		opt(&o)
	}
	if o.disable {
		return googlesql.Init(googlesql.WithCompilationMode(googlesql.CompilationModeInterpreter))
	}
	dir, err := Resolve(o.dir)
	if err != nil {
		return err
	}
	dir, err = EnsureSafe(dir)
	if err != nil {
		return err
	}
	return googlesql.Init(
		googlesql.WithCompilationMode(o.compileMode),
		googlesql.WithCompilationCache(dir),
	)
}

// linkedVersion returns the version of dependencyModulePath as
// recorded in this binary's build info. Falls back to "unversioned"
// when build info is missing (e.g. running from a `go run` of a
// non-module file or in an unusual build context).
func linkedVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "unversioned"
	}
	for _, dep := range bi.Deps {
		if dep == nil {
			continue
		}
		if dep.Path == dependencyModulePath && dep.Version != "" {
			return sanitizeVersion(dep.Version)
		}
	}
	return "unversioned"
}

// sanitizeVersion converts a module version into a directory-safe
// segment. Module versions are already simple (vX.Y.Z[-pre][+meta]
// or pseudo-versions), so we mostly need to make sure no path
// separators sneak in.
func sanitizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "unversioned"
	}
	v = filepath.Base(v)
	for _, c := range v {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '.' || c == '-' || c == '_' || c == '+':
		default:
			return "unversioned"
		}
	}
	return v
}

func isUnder(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return false
	}
	return true
}
