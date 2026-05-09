// Package runtime is an internal helper that ensures
// goccy/go-googlesql is initialised exactly once for callers that do
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
// googlesql.Init itself is sync.Once-guarded, but we wrap it here
// so callers that race Ensure() get a stable result.
func Ensure(opts ...googlesql.Option) error {
	once.Do(func() {
		initErr = googlesql.Init(opts...)
	})
	return initErr
}
