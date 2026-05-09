package executequery

import (
	"fmt"

	googlesql "github.com/goccy/go-googlesql"
)

// Config carries the user's effective request to Run. The exported
// fields mirror the upstream ExecuteQueryConfig surface plus the
// flag-rejection slots required by the unsupported flags (CLI layer
// uses these as sinks).
type Config struct {
	// Modes default to []Mode{ModeAnalyze} when empty.
	Modes []Mode

	// SQLMode defaults to SQLModeQuery.
	SQLMode SQLMode

	// CatalogName selects the active catalog (none / sample / tpch).
	// Empty defaults to "none". The catalog is built by package
	// catalog and passed to AnalyzeStatement; CatalogName here is
	// just the chosen identifier so options.go remains decoupled
	// from the catalog package.
	CatalogName string

	// LanguageOptions inputs.
	ProductMode               ProductMode
	StrictNameResolutionMode  bool
	EnabledLanguageFeatures   FeatureSet
	SupportsAllStatementKinds bool // default true; set false to honour upstream's per-mode kind set

	// AnalyzerOptions inputs.
	FoldLiteralCast           *bool
	PruneUnusedColumns        *bool
	ParseLocationRecordType   ParseLocationRecordType
	EnabledASTRewrites        RewriteSet
	Parameters                []QueryParameter
	AllowUndeclaredParameters bool

	// Unsupported sinks. The CLI layer populates these so Run can
	// surface a structured ErrUnsupportedFlag at validation time.
	ImportPaths                         []string
	TableSpecs                          []string
	DescriptorPool                      string
	TargetSyntax                        string
	UseBoxGlyphs                        *bool
	OutputMode                          string
	EvaluatorMaxValueByteSize           *int64
	EvaluatorMaxIntermediateByteSize    *int64
	EvaluatorScrambleUndefinedOrderings *bool
	MaxStatementsToExecute              *int64
	Web                                 bool
	Port                                *int
}

// Validate inspects cfg for unsupported flag values, returning the
// first ErrUnsupported* (wrapped) error found. The order is stable
// so test goldens can pin it.
func (c *Config) Validate() error {
	for _, m := range c.Modes {
		if !m.IsSupported() {
			return modeUnsupportedf(string(m), m.UnsupportedReason())
		}
	}
	if c.SQLMode != "" && c.SQLMode != SQLModeQuery && c.SQLMode != SQLModeExpression && c.SQLMode != SQLModeScript {
		return sqlModeUnsupportedf(string(c.SQLMode), "unknown sql_mode")
	}

	if c.TargetSyntax != "" && c.TargetSyntax != "standard" {
		return FlagUnsupportedError("target_syntax", ReasonFlagTargetSyntax)
	}
	if c.UseBoxGlyphs != nil {
		return FlagUnsupportedError("use_box_glyphs", ReasonFlagUseBoxGlyphs)
	}
	if c.OutputMode != "" && c.OutputMode != "box" {
		return FlagUnsupportedError("output_mode", ReasonFlagOutputMode)
	}
	if len(c.TableSpecs) != 0 {
		return FlagUnsupportedError("table_spec", ReasonFlagTableSpec)
	}
	if c.DescriptorPool != "" && c.DescriptorPool != "none" {
		return FlagUnsupportedError("descriptor_pool", ReasonFlagDescriptorPool)
	}
	if c.EvaluatorMaxValueByteSize != nil {
		return FlagUnsupportedError("evaluator_max_value_byte_size", ReasonFlagEvaluatorMaxValue)
	}
	if c.EvaluatorMaxIntermediateByteSize != nil {
		return FlagUnsupportedError("evaluator_max_intermediate_byte_size", ReasonFlagEvaluatorMaxIntermediate)
	}
	if c.EvaluatorScrambleUndefinedOrderings != nil {
		return FlagUnsupportedError("evaluator_scramble_undefined_orderings", ReasonFlagEvaluatorScramble)
	}
	if c.MaxStatementsToExecute != nil {
		return FlagUnsupportedError("max_statements_to_execute", ReasonFlagMaxStatementsToExecute)
	}
	if len(c.ImportPaths) != 0 {
		return FlagUnsupportedError("import_path", ReasonFlagImportPath)
	}
	if c.Web {
		return FlagUnsupportedError("web", ReasonFlagWeb)
	}
	if c.Port != nil {
		return FlagUnsupportedError("port", ReasonFlagPort)
	}
	return nil
}

// effectiveModes returns Modes with the default applied.
func (c *Config) effectiveModes() []Mode {
	if len(c.Modes) == 0 {
		return []Mode{ModeAnalyze}
	}
	return c.Modes
}

// effectiveSQLMode returns SQLMode with the default applied.
func (c *Config) effectiveSQLMode() SQLMode {
	if c.SQLMode == "" {
		return SQLModeQuery
	}
	return c.SQLMode
}

// buildLanguageOptions translates the LanguageOptions-bound fields
// of c into a freshly-constructed *googlesql.LanguageOptions.
func (c *Config) buildLanguageOptions() (*googlesql.LanguageOptions, error) {
	lo, err := googlesql.NewLanguageOptions()
	if err != nil {
		return nil, fmt.Errorf("new language options: %w", err)
	}

	// Default to upstream's ALL_MINUS_DEV when no explicit base is
	// set: that matches `--enabled_ast_rewrites`'s upstream default
	// and gives sensible analyzer behaviour.
	if c.EnabledLanguageFeatures.Base == FeatureBaseUnset && len(c.EnabledLanguageFeatures.Enabled) == 0 && len(c.EnabledLanguageFeatures.Disabled) == 0 {
		if err := lo.EnableMaximumLanguageFeatures(); err != nil {
			return nil, fmt.Errorf("enable maximum features: %w", err)
		}
	} else if err := c.EnabledLanguageFeatures.Apply(lo); err != nil {
		return nil, fmt.Errorf("language features: %w", err)
	}

	if c.SupportsAllStatementKinds || (c.Modes == nil) {
		// Default to "all kinds" so DESCRIBE / DDL parse paths work
		// without per-mode kind whitelisting.
		if err := lo.SetSupportsAllStatementKinds(); err != nil {
			return nil, fmt.Errorf("set supports all statement kinds: %w", err)
		}
	}

	switch c.ProductMode {
	case "", ProductModeInternal:
		if err := lo.SetProductMode(googlesql.ProductModeProductInternal); err != nil {
			return nil, fmt.Errorf("set product_mode: %w", err)
		}
	case ProductModeExternal:
		if err := lo.SetProductMode(googlesql.ProductModeProductExternal); err != nil {
			return nil, fmt.Errorf("set product_mode: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown product_mode %q", c.ProductMode)
	}

	if c.StrictNameResolutionMode {
		if err := lo.SetNameResolutionMode(googlesql.NameResolutionModeNameResolutionStrict); err != nil {
			return nil, fmt.Errorf("set name_resolution_mode: %w", err)
		}
	}

	return lo, nil
}

// buildAnalyzerOptions translates the AnalyzerOptions-bound fields
// of c into a freshly-constructed *googlesql.AnalyzerOptions, with
// the linked LanguageOptions installed.
func (c *Config) buildAnalyzerOptions(lo *googlesql.LanguageOptions, tf *googlesql.TypeFactory) (*googlesql.AnalyzerOptions, error) {
	ao, err := googlesql.NewAnalyzerOptions2()
	if err != nil {
		return nil, fmt.Errorf("new analyzer options: %w", err)
	}
	if err := ao.SetLanguage(lo); err != nil {
		return nil, fmt.Errorf("set language: %w", err)
	}
	if c.FoldLiteralCast != nil {
		if err := ao.SetFoldLiteralCast(*c.FoldLiteralCast); err != nil {
			return nil, fmt.Errorf("set fold_literal_cast: %w", err)
		}
	}
	if c.PruneUnusedColumns != nil {
		if err := ao.SetPruneUnusedColumns(*c.PruneUnusedColumns); err != nil {
			return nil, fmt.Errorf("set prune_unused_columns: %w", err)
		}
	}
	if c.ParseLocationRecordType != "" {
		if err := ao.SetParseLocationRecordType(parseLocationRecordToGoogleSQL(c.ParseLocationRecordType)); err != nil {
			return nil, fmt.Errorf("set parse_location_record_type: %w", err)
		}
	}
	if err := c.EnabledASTRewrites.Apply(ao); err != nil {
		return nil, fmt.Errorf("ast rewrites: %w", err)
	}
	if c.AllowUndeclaredParameters {
		if err := ao.SetAllowUndeclaredParameters(true); err != nil {
			return nil, fmt.Errorf("set allow_undeclared_parameters: %w", err)
		}
	}
	for _, p := range c.Parameters {
		typ, err := tf.MakeSimpleType(p.Type)
		if err != nil {
			return nil, fmt.Errorf("make type for parameter %q: %w", p.Name, err)
		}
		if err := ao.AddQueryParameter(p.Name, typ); err != nil {
			return nil, fmt.Errorf("add parameter %q: %w", p.Name, err)
		}
	}
	return ao, nil
}

func parseLocationRecordToGoogleSQL(t ParseLocationRecordType) googlesql.ParseLocationRecordType {
	switch t {
	case ParseLocationRecordFullNodeScope:
		return googlesql.ParseLocationRecordTypeParseLocationRecordFullNodeScope
	case ParseLocationRecordCodeSearch:
		return googlesql.ParseLocationRecordTypeParseLocationRecordCodeSearch
	}
	return googlesql.ParseLocationRecordTypeParseLocationRecordNone
}
