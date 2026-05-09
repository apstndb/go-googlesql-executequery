package executequery

import (
	"fmt"
	"strings"
	"sync"

	googlesql "github.com/goccy/go-googlesql"
)

// FeatureSet is the parsed value of `--enabled_language_features`.
//
// Upstream's flag format is BASE[,+FOO][,-BAR]:
//
//	BASE   ∈ {NONE, ALL, ALL_MINUS_DEV, DEFAULTS, DEFAULTS_MINUS_DEV}
//	+FOO   enable feature FOO on top of BASE
//	-BAR   disable feature BAR from BASE
//
// Feature names are upstream's enum spelling (e.g. FEATURE_RANGE_TYPE
// or FEATURE_V_1_1_WITH_ON_SUBQUERY). Names are matched
// case-insensitively and underscore-insensitively against the Go
// binding's LanguageFeature enum, which collapses runs like "V_1_1"
// to "V11" — both forms are accepted.
type FeatureSet struct {
	Base     FeatureBase
	Enabled  []googlesql.LanguageFeature
	Disabled []googlesql.LanguageFeature
}

// FeatureBase mirrors upstream's set-base keyword.
type FeatureBase int

const (
	// FeatureBaseUnset means no base keyword was provided. Apply only
	// the +/- modifiers.
	FeatureBaseUnset FeatureBase = iota

	// FeatureBaseNone disables all features.
	FeatureBaseNone

	// FeatureBaseAll enables all features (including in-development).
	// Maps to LanguageOptions.EnableMaximumLanguageFeaturesForDevelopment.
	FeatureBaseAll

	// FeatureBaseAllMinusDev enables non-in-development features.
	// Maps to LanguageOptions.EnableMaximumLanguageFeatures.
	FeatureBaseAllMinusDev

	// FeatureBaseDefaults uses NewLanguageOptions's defaults
	// (no-op; the LanguageOptions starts in the defaults state).
	//
	// NOTE: upstream's DEFAULTS is computed from a separate
	// classification metadata. goccy/go-googlesql does not expose
	// that classification, so we treat DEFAULTS as the
	// NewLanguageOptions zero state. This matches in-practice but
	// may diverge from upstream for fringe features.
	FeatureBaseDefaults

	// FeatureBaseDefaultsMinusDev — see FeatureBaseDefaults caveat.
	FeatureBaseDefaultsMinusDev
)

// ParseFeatureSet parses the `--enabled_language_features` flag.
func ParseFeatureSet(s string) (FeatureSet, error) {
	var fs FeatureSet
	s = strings.TrimSpace(s)
	if s == "" {
		return fs, nil
	}
	for i, raw := range strings.Split(s, ",") {
		tok := strings.TrimSpace(raw)
		if tok == "" {
			continue
		}
		switch {
		case strings.HasPrefix(tok, "+"):
			feature, err := lookupLanguageFeature(tok[1:])
			if err != nil {
				return FeatureSet{}, fmt.Errorf("enabled_language_features: %q: %w", raw, err)
			}
			fs.Enabled = append(fs.Enabled, feature)
		case strings.HasPrefix(tok, "-"):
			feature, err := lookupLanguageFeature(tok[1:])
			if err != nil {
				return FeatureSet{}, fmt.Errorf("enabled_language_features: %q: %w", raw, err)
			}
			fs.Disabled = append(fs.Disabled, feature)
		default:
			if i != 0 {
				return FeatureSet{}, fmt.Errorf("enabled_language_features: base %q must be the first token", raw)
			}
			base, err := parseFeatureBase(tok)
			if err != nil {
				return FeatureSet{}, fmt.Errorf("enabled_language_features: %w", err)
			}
			fs.Base = base
		}
	}
	return fs, nil
}

// Apply mutates lo to reflect fs.
func (fs FeatureSet) Apply(lo *googlesql.LanguageOptions) error {
	switch fs.Base {
	case FeatureBaseUnset, FeatureBaseDefaults, FeatureBaseDefaultsMinusDev:
		// no-op (start from NewLanguageOptions's defaults)
	case FeatureBaseNone:
		if err := lo.DisableAllLanguageFeatures(); err != nil {
			return fmt.Errorf("disable all features: %w", err)
		}
	case FeatureBaseAllMinusDev:
		if err := lo.EnableMaximumLanguageFeatures(); err != nil {
			return fmt.Errorf("enable maximum features: %w", err)
		}
	case FeatureBaseAll:
		if err := lo.EnableMaximumLanguageFeaturesForDevelopment(); err != nil {
			return fmt.Errorf("enable maximum features (dev): %w", err)
		}
	}
	for _, f := range fs.Enabled {
		if err := lo.EnableLanguageFeature(f); err != nil {
			return fmt.Errorf("enable %s: %w", f, err)
		}
	}
	for _, f := range fs.Disabled {
		if err := lo.DisableLanguageFeature(f); err != nil {
			return fmt.Errorf("disable %s: %w", f, err)
		}
	}
	return nil
}

func parseFeatureBase(s string) (FeatureBase, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "NONE":
		return FeatureBaseNone, nil
	case "ALL":
		return FeatureBaseAll, nil
	case "ALL_MINUS_DEV":
		return FeatureBaseAllMinusDev, nil
	case "DEFAULTS":
		return FeatureBaseDefaults, nil
	case "DEFAULTS_MINUS_DEV":
		return FeatureBaseDefaultsMinusDev, nil
	}
	return 0, fmt.Errorf("unknown base %q (expected NONE | ALL | ALL_MINUS_DEV | DEFAULTS | DEFAULTS_MINUS_DEV)", s)
}

// lookupLanguageFeature finds a LanguageFeature by upstream-style
// name (e.g. FEATURE_RANGE_TYPE). Underscores are ignored, matching
// is case-insensitive.
func lookupLanguageFeature(name string) (googlesql.LanguageFeature, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("empty feature name")
	}
	m := languageFeatureMap()
	if f, ok := m[normalizeFeatureName(name)]; ok {
		return f, nil
	}
	return 0, fmt.Errorf("unknown LanguageFeature %q", name)
}

func normalizeFeatureName(s string) string {
	// Strip underscores and uppercase. The Go binding's enum names
	// follow the form "LanguageFeatureFeature<CamelCase>", and
	// upstream's names are "FEATURE_<SCREAMING_SNAKE>". Both
	// collapse to the same underscore-free uppercase form.
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			continue
		}
		if c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		b.WriteByte(c)
	}
	return b.String()
}

var (
	langFeatureMapOnce sync.Once
	langFeatureMap     map[string]googlesql.LanguageFeature
)

func languageFeatureMap() map[string]googlesql.LanguageFeature {
	langFeatureMapOnce.Do(func() {
		langFeatureMap = make(map[string]googlesql.LanguageFeature, len(allLanguageFeatures))
		for _, f := range allLanguageFeatures {
			langFeatureMap[normalizeFeatureName(strings.TrimPrefix(f.String(), "LanguageFeature"))] = f
		}
	})
	return langFeatureMap
}
