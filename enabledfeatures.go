package executequery

import (
	"fmt"
	"strings"
	"sync"

	"github.com/apstndb/go-googlesql-executequery/internal/optionspb"
	googlesql "github.com/goccy/go-googlesql"
	"google.golang.org/protobuf/proto"
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

	// FeatureBaseDefaults uses NewLanguageOptions's defaults.
	FeatureBaseDefaults

	// FeatureBaseDefaultsMinusDev uses NewLanguageOptions's defaults
	// minus features that are in development.
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
	case FeatureBaseUnset, FeatureBaseDefaults:
		// no-op (start from NewLanguageOptions's defaults)
	case FeatureBaseDefaultsMinusDev:
		for _, f := range allLanguageFeatures {
			if isLanguageFeatureInDevelopment(f) {
				if err := lo.DisableLanguageFeature(f); err != nil {
					return fmt.Errorf("disable dev feature %s: %w", f, err)
				}
			}
		}
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

// normalizeFeatureName collapses both go-googlesql's Go-style enum
// name and upstream's `FEATURE_<SCREAMING_SNAKE>` flag spelling to
// a common underscore-free upper-case key, so the same lookup map
// matches whichever form the user types.
func normalizeFeatureName(s string) string {
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

// isLanguageFeatureInDevelopment checks the `in_development` proto
// annotation for a given go-googlesql LanguageFeature value.
//
// Because the submodule proto and go-googlesql may be built from
// different commits of options.proto (enum numbers can shift between
// versions), we match by **name**, not by numeric value.
//
// Implementation note: the lookup builds a cached name→annotation
// map from the proto on first call, then matches against the
// go-googlesql enum's String() representation.
func isLanguageFeatureInDevelopment(f googlesql.LanguageFeature) bool {
	m := langFeatureAnnotations()
	// go-googlesql's String() returns e.g. "LanguageFeatureFeatureRowType";
	// strip the "LanguageFeature" prefix and normalize to match proto names.
	key := normalizeFeatureName(strings.TrimPrefix(f.String(), "LanguageFeature"))
	if ann, ok := m[key]; ok {
		return ann.GetInDevelopment()
	}
	return false
}

var (
	langFeatureAnnotationsOnce sync.Once
	langFeatureAnnotationsMap  map[string]*optionspb.LanguageFeatureOptions
)

// langFeatureAnnotations returns a normalized-name → LanguageFeatureOptions
// map built from the proto enum value options.
func langFeatureAnnotations() map[string]*optionspb.LanguageFeatureOptions {
	langFeatureAnnotationsOnce.Do(func() {
		ed := optionspb.File_googlesql_public_options_proto.Enums().ByName("LanguageFeature")
		langFeatureAnnotationsMap = make(map[string]*optionspb.LanguageFeatureOptions, ed.Values().Len())
		for i := 0; i < ed.Values().Len(); i++ {
			vd := ed.Values().Get(i)
			opts := vd.Options()
			if !proto.HasExtension(opts, optionspb.E_LanguageFeatureOptions) {
				continue
			}
			ext := proto.GetExtension(opts, optionspb.E_LanguageFeatureOptions).(*optionspb.LanguageFeatureOptions)
			// Register under normalized proto name for matching against
			// go-googlesql's Go-style enum name (both normalize the same).
			langFeatureAnnotationsMap[normalizeFeatureName(string(vd.Name()))] = ext
		}
	})
	return langFeatureAnnotationsMap
}
