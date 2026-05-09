// Package runtime is an internal helper that ensures
// go-googlesql is initialised exactly once for callers that do
// not go through cache.Setup directly.
package runtime

import (
	"sync"

	googlesql "github.com/goccy/go-googlesql"
)

var (
	once    sync.Once
	initErr error
)

// Ensure calls googlesql.Init(opts...) exactly once for this
// process. The first caller's options win; subsequent calls return
// the first call's error (or nil) without invoking Init again.
//
// Workaround for go-googlesql v0.2.1: `googlesql.Init` is itself
// `sync.Once`-guarded, but its return value on second-and-later
// calls is nil regardless of whether the first call failed, so a
// caller that races Init cannot observe a stable error.
//
// Upstream C++ API: none — `googlesql.Init` is a Go-binding-only
// entry point that drives wazero/wasm runtime startup; there is no
// equivalent symbol in `google/googlesql` (the C++ library is just
// linked, not "initialised"). Treat this as a binding-shape bug
// rather than a missing C++ accessor.
//
// Natural Go code:
//
//	if err := googlesql.Init(opts...); err != nil { ... }   // call as many times as you want
//
// Instead, the local `sync.Once` here remembers initErr so every
// caller sees the same outcome. Unblocked when go-googlesql either
// makes Init's "already initialised" return path replay the original
// error, or exposes a separate `googlesql.InitError()` accessor.
func Ensure(opts ...googlesql.Option) error {
	once.Do(func() {
		initErr = googlesql.Init(opts...)
	})
	return initErr
}
