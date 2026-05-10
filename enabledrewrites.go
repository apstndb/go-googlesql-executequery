package executequery

import (
	"fmt"
	"strings"
	"sync"

	"github.com/apstndb/go-googlesql-executequery/internal/optionspb"
	googlesql "github.com/goccy/go-googlesql"
	"google.golang.org/protobuf/proto"
)

// RewriteSet is the parsed value of `--enabled_ast_rewrites`.
//
// Upstream's flag format mirrors --enabled_language_features:
// BASE[,+REWRITE_FOO][,-REWRITE_BAR]. The default upstream base is
// ALL_MINUS_DEV.
type RewriteSet struct {
	Base     RewriteBase
	Enabled  []googlesql.ResolvedASTRewrite
	Disabled []googlesql.ResolvedASTRewrite
}

// RewriteBase mirrors upstream's set-base keyword.
type RewriteBase int

const (
	// RewriteBaseUnset means no base; apply only +/- modifiers.
	RewriteBaseUnset RewriteBase = iota

	// RewriteBaseNone disables all rewrites.
	RewriteBaseNone

	// RewriteBaseAll enables all rewrites (including in-development).
	RewriteBaseAll

	// RewriteBaseAllMinusDev enables non-in-development rewrites.
	// (Default upstream behaviour.)
	RewriteBaseAllMinusDev

	// RewriteBaseDefaults uses NewAnalyzerOptions's defaults.
	//
	// Workaround [go-googlesql v0.2.1]: the static helper that
	// returns upstream's DEFAULTS rewrite set is not exposed.
	//
	// Upstream C++ API:
	// googlesql::AnalyzerOptions::DefaultRewrites() (static, returning
	// `absl::btree_set<ResolvedASTRewrite>`) at
	// third_party/googlesql/googlesql/public/analyzer_options.h:342.
	// `enabled_rewrites = DefaultRewrites()` is what
	// `NewAnalyzerOptions()` initialises with (analyzer_options.h:1067),
	// so the post-construction state is already DEFAULTS — but we
	// can't *recompute* the set later if the user mixes DEFAULTS with
	// `+REWRITE_FOO` / `-REWRITE_BAR` modifiers.
	//
	// Natural Go code:
	//   ao.SetEnabledRewrites(googlesql.DefaultRewrites())
	//
	// Instead, DEFAULTS is treated as the NewAnalyzerOptions zero
	// state. Unblocked when go-googlesql exposes
	// `AnalyzerOptions.DefaultRewrites` (or an equivalent static
	// accessor).
	RewriteBaseDefaults

	// RewriteBaseDefaultsMinusDev — see RewriteBaseDefaults caveat.
	RewriteBaseDefaultsMinusDev
)

// ParseRewriteSet parses the `--enabled_ast_rewrites` flag.
func ParseRewriteSet(s string) (RewriteSet, error) {
	var rs RewriteSet
	s = strings.TrimSpace(s)
	if s == "" {
		return rs, nil
	}
	for i, raw := range strings.Split(s, ",") {
		tok := strings.TrimSpace(raw)
		if tok == "" {
			continue
		}
		switch {
		case strings.HasPrefix(tok, "+"):
			rw, err := lookupRewrite(tok[1:])
			if err != nil {
				return RewriteSet{}, fmt.Errorf("enabled_ast_rewrites: %q: %w", raw, err)
			}
			rs.Enabled = append(rs.Enabled, rw)
		case strings.HasPrefix(tok, "-"):
			rw, err := lookupRewrite(tok[1:])
			if err != nil {
				return RewriteSet{}, fmt.Errorf("enabled_ast_rewrites: %q: %w", raw, err)
			}
			rs.Disabled = append(rs.Disabled, rw)
		default:
			if i != 0 {
				return RewriteSet{}, fmt.Errorf("enabled_ast_rewrites: base %q must be the first token", raw)
			}
			base, err := parseRewriteBase(tok)
			if err != nil {
				return RewriteSet{}, fmt.Errorf("enabled_ast_rewrites: %w", err)
			}
			rs.Base = base
		}
	}
	return rs, nil
}

// Apply mutates ao to reflect rs. Note: setting BASE iterates every
// known rewrite and toggles it explicitly; this matches upstream
// (which builds the effective set the same way).
func (rs RewriteSet) Apply(ao *googlesql.AnalyzerOptions) error {
	switch rs.Base {
	case RewriteBaseUnset, RewriteBaseDefaults, RewriteBaseDefaultsMinusDev:
		// no-op (start from NewAnalyzerOptions's defaults)
	case RewriteBaseNone:
		for _, rw := range allResolvedASTRewrites {
			if err := ao.EnableRewrite(rw, false); err != nil {
				return fmt.Errorf("disable %s: %w", rw, err)
			}
		}
	case RewriteBaseAll, RewriteBaseAllMinusDev:
		for _, rw := range allResolvedASTRewrites {
			if rs.Base == RewriteBaseAllMinusDev && isRewriteInDevelopment(rw) {
				continue
			}
			if err := ao.EnableRewrite(rw, true); err != nil {
				return fmt.Errorf("enable %s: %w", rw, err)
			}
		}
	}
	for _, rw := range rs.Enabled {
		if err := ao.EnableRewrite(rw, true); err != nil {
			return fmt.Errorf("enable %s: %w", rw, err)
		}
	}
	for _, rw := range rs.Disabled {
		if err := ao.EnableRewrite(rw, false); err != nil {
			return fmt.Errorf("disable %s: %w", rw, err)
		}
	}
	return nil
}

func parseRewriteBase(s string) (RewriteBase, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "NONE":
		return RewriteBaseNone, nil
	case "ALL":
		return RewriteBaseAll, nil
	case "ALL_MINUS_DEV":
		return RewriteBaseAllMinusDev, nil
	case "DEFAULTS":
		return RewriteBaseDefaults, nil
	case "DEFAULTS_MINUS_DEV":
		return RewriteBaseDefaultsMinusDev, nil
	}
	return 0, fmt.Errorf("unknown base %q (expected NONE | ALL | ALL_MINUS_DEV | DEFAULTS | DEFAULTS_MINUS_DEV)", s)
}

func lookupRewrite(name string) (googlesql.ResolvedASTRewrite, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("empty rewrite name")
	}
	m := rewriteMap()
	if rw, ok := m[normalizeFeatureName(name)]; ok {
		return rw, nil
	}
	return 0, fmt.Errorf("unknown ResolvedASTRewrite %q", name)
}

var (
	rewriteMapOnce sync.Once
	rewriteMapVal  map[string]googlesql.ResolvedASTRewrite
)

// rewriteMap caches the user-facing-name → ResolvedASTRewrite
// lookup used by ParseRewriteSet. Uses go-googlesql's enum values
// (from zz_enums.go) because the wasm binding assigns different
// numeric values than the committed proto.
func rewriteMap() map[string]googlesql.ResolvedASTRewrite {
	rewriteMapOnce.Do(func() {
		rewriteMapVal = make(map[string]googlesql.ResolvedASTRewrite, len(allResolvedASTRewrites))
		for _, rw := range allResolvedASTRewrites {
			name := strings.TrimPrefix(rw.String(), "ResolvedASTRewrite")
			name = strings.TrimPrefix(name, "Rewrite")
			rewriteMapVal[normalizeFeatureName(name)] = rw
		}
	})
	return rewriteMapVal
}

// isRewriteInDevelopment checks the `in_development` proto annotation
// for a given go-googlesql ResolvedASTRewrite value.
//
// Because the wasm binding assigns different numeric enum values than
// the committed proto (systematic +1 offset observed in go-googlesql
// v0.2.1), we match by **name**, not by numeric value.
func isRewriteInDevelopment(rw googlesql.ResolvedASTRewrite) bool {
	m := rewriteAnnotations()
	// go-googlesql's String() returns e.g. "ResolvedASTRewriteRewriteFlatten";
	// strip "ResolvedASTRewrite" prefix, then "Rewrite" prefix, and normalize.
	key := strings.TrimPrefix(rw.String(), "ResolvedASTRewrite")
	key = strings.TrimPrefix(key, "Rewrite")
	key = normalizeFeatureName(key)
	if ann, ok := m[key]; ok {
		return ann.GetInDevelopment()
	}
	return false
}

var (
	rewriteAnnotationsOnce sync.Once
	rewriteAnnotationsMap  map[string]*optionspb.ResolvedASTRewriteOptions
)

// rewriteAnnotations returns a normalized-name → ResolvedASTRewriteOptions
// map built from the proto enum value options.
func rewriteAnnotations() map[string]*optionspb.ResolvedASTRewriteOptions {
	rewriteAnnotationsOnce.Do(func() {
		ed := optionspb.File_googlesql_public_options_proto.Enums().ByName("ResolvedASTRewrite")
		rewriteAnnotationsMap = make(map[string]*optionspb.ResolvedASTRewriteOptions, ed.Values().Len())
		for i := 0; i < ed.Values().Len(); i++ {
			vd := ed.Values().Get(i)
			opts := vd.Options()
			if !proto.HasExtension(opts, optionspb.E_RewriteOptions) {
				continue
			}
			ext := proto.GetExtension(opts, optionspb.E_RewriteOptions).(*optionspb.ResolvedASTRewriteOptions)
			// Register under normalized proto name (strip "REWRITE_" prefix).
			name := strings.TrimPrefix(string(vd.Name()), "REWRITE_")
			rewriteAnnotationsMap[normalizeFeatureName(name)] = ext
		}
	})
	return rewriteAnnotationsMap
}
