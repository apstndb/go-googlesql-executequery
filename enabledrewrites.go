package executequery

import (
	"fmt"
	"strings"
	"sync"

	googlesql "github.com/goccy/go-googlesql"
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
	// Workaround for goccy/go-googlesql v0.2.1: upstream's DEFAULTS
	// reads each rewrite's `default_enabled` flag, which goccy does
	// not expose. See FeatureBaseDefaults for the same shape.
	//
	// Natural code:
	//   for _, rw := range googlesql.AllResolvedASTRewrites() {
	//       if rw.IsDefaultEnabled() { ao.EnableRewrite(rw, true) }
	//   }
	//
	// Instead, DEFAULTS is treated as the NewAnalyzerOptions zero
	// state. Unblocked alongside FeatureBaseDefaults.
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
		// Workaround for goccy/go-googlesql v0.2.1: rewrites are not
		// classified as in-development vs general-availability, so
		// ALL and ALL_MINUS_DEV both enable everything we know about.
		//
		// Natural code:
		//   for _, rw := range googlesql.AllResolvedASTRewrites() {
		//       if base == RewriteBaseAllMinusDev && rw.IsInDevelopment() {
		//           continue
		//       }
		//       ao.EnableRewrite(rw, true)
		//   }
		//
		// Unblocked when goccy exports an `IsInDevelopment()`
		// classifier on `ResolvedASTRewrite`.
		for _, rw := range allResolvedASTRewrites {
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

func rewriteMap() map[string]googlesql.ResolvedASTRewrite {
	rewriteMapOnce.Do(func() {
		rewriteMapVal = make(map[string]googlesql.ResolvedASTRewrite, len(allResolvedASTRewrites))
		for _, rw := range allResolvedASTRewrites {
			// String() returns "ResolvedASTRewriteRewriteFoo"; strip
			// the outer "ResolvedASTRewrite" prefix and the inner
			// "Rewrite" prefix so the user-facing name matches
			// upstream's "FOO" form (upstream's --enabled_ast_rewrites
			// help: "Enum values must be listed with 'REWRITE_'
			// stripped").
			name := strings.TrimPrefix(rw.String(), "ResolvedASTRewrite")
			name = strings.TrimPrefix(name, "Rewrite")
			rewriteMapVal[normalizeFeatureName(name)] = rw
		}
	})
	return rewriteMapVal
}
