package main

import (
	"flag"
	"fmt"
	"strings"

	executequery "github.com/apstndb/go-googlesql-executequery"
)

// registeredFlags holds the raw values of every CLI flag, plus
// per-flag "was set?" sentinels for boolean and integer flags whose
// default value is meaningful in upstream and should not on its own
// trigger an unsupported-flag error.
type registeredFlags struct {
	// Supported flags.
	mode                     string
	sqlMode                  string
	catalog                  string
	productMode              string
	strictNameResolutionMode bool
	enabledLanguageFeatures  string
	enabledASTRewrites       string
	foldLiteralCast          string // empty == unset; "true"/"false" otherwise
	pruneUnusedColumns       string
	parseLocationRecordType  string
	parameters               string

	// Unsupported sinks (see executequery.ReasonFlag* for why).
	importPath                          string
	tableSpec                           string
	descriptorPool                      string
	targetSyntax                        string
	useBoxGlyphsExplicit                bool
	useBoxGlyphs                        bool
	outputMode                          string
	evaluatorMaxValueByteSize           int64
	evaluatorMaxValueByteSizeExplicit   bool
	evaluatorMaxIntermediateByteSize    int64
	evaluatorMaxIntermediateExplicit    bool
	evaluatorScrambleUndefinedOrderings bool
	evaluatorScrambleExplicit           bool
	maxStatementsToExecute              int64
	maxStatementsExplicit               bool
	web                                 bool
	port                                int
	portExplicit                        bool

	// Go-port-specific.
	cacheDir        string
	noCache         bool
	compilationMode string
	sqlFile         string
}

// registerFlags wires every upstream `execute_query` flag, plus a
// small set of Go-port-specific flags, into fs.
func registerFlags(fs *flag.FlagSet) *registeredFlags {
	// Defaults mirror upstream where the value is implementable; for
	// upstream defaults that this Go port cannot honour (e.g.
	// --descriptor_pool=generated), the flag default is left empty so
	// "user did not set it" is distinguishable from "user set it to
	// upstream's default" — only the latter triggers
	// ErrUnsupportedFlag.
	rf := &registeredFlags{
		mode:                    "analyze",
		sqlMode:                 "query",
		catalog:                 "none",
		productMode:             "internal",
		parseLocationRecordType: "NONE",
		compilationMode:         "compiler",
	}

	// --- Supported flags ---

	fs.StringVar(&rf.mode, "mode", rf.mode, helpMode)
	fs.StringVar(&rf.sqlMode, "sql_mode", rf.sqlMode, helpSQLMode)
	fs.StringVar(&rf.catalog, "catalog", rf.catalog, helpCatalog)
	fs.StringVar(&rf.productMode, "product_mode", rf.productMode, helpProductMode)
	fs.BoolVar(&rf.strictNameResolutionMode, "strict_name_resolution_mode", false, helpStrictNameResolutionMode)
	fs.StringVar(&rf.enabledLanguageFeatures, "enabled_language_features", "", helpEnabledLanguageFeatures)
	fs.StringVar(&rf.enabledASTRewrites, "enabled_ast_rewrites", "", helpEnabledASTRewrites)
	// fold_literal_cast and prune_unused_columns default to true
	// upstream and unsetting them changes analyzer behaviour, so we
	// model them as tri-state strings ("" | "true" | "false") to
	// distinguish "unset" from "set to default".
	fs.Func("fold_literal_cast", helpFoldLiteralCast, func(v string) error {
		rf.foldLiteralCast = strings.ToLower(strings.TrimSpace(v))
		return nil
	})
	fs.Func("prune_unused_columns", helpPruneUnusedColumns, func(v string) error {
		rf.pruneUnusedColumns = strings.ToLower(strings.TrimSpace(v))
		return nil
	})
	fs.StringVar(&rf.parseLocationRecordType, "parse_location_record_type", rf.parseLocationRecordType, helpParseLocationRecordType)
	fs.StringVar(&rf.parameters, "parameters", "", helpParameters)

	// --- Unsupported sinks ---
	// Each unsupported flag carries the upstream blocker (ReasonFlag*)
	// in its help suffix and triggers an unsupported error from
	// toConfig() when set to a non-default value.

	fs.StringVar(&rf.importPath, "import_path", "", unsupportedHelp(helpImportPath, executequery.ReasonFlagImportPath))
	fs.StringVar(&rf.tableSpec, "table_spec", "", unsupportedHelp(helpTableSpec, executequery.ReasonFlagTableSpec))
	// descriptor_pool / target_syntax / output_mode default to ""
	// (unset). Validation accepts "" and "none"/standard"/"box" silently
	// and rejects any other explicit value with ErrUnsupportedFlag.
	fs.StringVar(&rf.descriptorPool, "descriptor_pool", "", unsupportedHelp(helpDescriptorPool, executequery.ReasonFlagDescriptorPool))
	fs.StringVar(&rf.targetSyntax, "target_syntax", "", unsupportedHelp(helpTargetSyntax, executequery.ReasonFlagTargetSyntax))
	fs.Func("use_box_glyphs", unsupportedHelp(helpUseBoxGlyphs, executequery.ReasonFlagUseBoxGlyphs), func(v string) error {
		rf.useBoxGlyphsExplicit = true
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1":
			rf.useBoxGlyphs = true
		case "false", "0":
			rf.useBoxGlyphs = false
		default:
			return fmt.Errorf("expected true|false")
		}
		return nil
	})
	fs.StringVar(&rf.outputMode, "output_mode", "", unsupportedHelp(helpOutputMode, executequery.ReasonFlagOutputMode))

	fs.Func("evaluator_max_value_byte_size", unsupportedHelp(helpEvaluatorMaxValueByteSize, executequery.ReasonFlagEvaluatorMaxValue), func(v string) error {
		rf.evaluatorMaxValueByteSizeExplicit = true
		_, err := fmt.Sscanf(v, "%d", &rf.evaluatorMaxValueByteSize)
		return err
	})
	fs.Func("evaluator_max_intermediate_byte_size", unsupportedHelp(helpEvaluatorMaxIntermediateByteSize, executequery.ReasonFlagEvaluatorMaxIntermediate), func(v string) error {
		rf.evaluatorMaxIntermediateExplicit = true
		_, err := fmt.Sscanf(v, "%d", &rf.evaluatorMaxIntermediateByteSize)
		return err
	})
	fs.Func("evaluator_scramble_undefined_orderings", unsupportedHelp(helpEvaluatorScrambleUndefinedOrderings, executequery.ReasonFlagEvaluatorScramble), func(v string) error {
		rf.evaluatorScrambleExplicit = true
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1":
			rf.evaluatorScrambleUndefinedOrderings = true
		case "false", "0":
			rf.evaluatorScrambleUndefinedOrderings = false
		default:
			return fmt.Errorf("expected true|false")
		}
		return nil
	})
	fs.Func("max_statements_to_execute", unsupportedHelp(helpMaxStatementsToExecute, executequery.ReasonFlagMaxStatementsToExecute), func(v string) error {
		rf.maxStatementsExplicit = true
		_, err := fmt.Sscanf(v, "%d", &rf.maxStatementsToExecute)
		return err
	})
	fs.BoolVar(&rf.web, "web", false, unsupportedHelp(helpWeb, executequery.ReasonFlagWeb))
	fs.Func("port", unsupportedHelp(helpPort, executequery.ReasonFlagPort), func(v string) error {
		rf.portExplicit = true
		_, err := fmt.Sscanf(v, "%d", &rf.port)
		return err
	})

	// --- Go-port-specific (not from upstream) ---

	fs.StringVar(&rf.cacheDir, "cache_dir", "", helpCacheDir)
	fs.BoolVar(&rf.noCache, "no_cache", false, helpNoCache)
	fs.StringVar(&rf.compilationMode, "compilation_mode", rf.compilationMode, helpCompilationMode)
	fs.StringVar(&rf.sqlFile, "sql_file", "", helpSQLFile)

	return rf
}

// unsupportedHelp annotates upstream help text with the Go-port
// rejection reason, so `--help` makes the gap obvious.
func unsupportedHelp(upstream, reason string) string {
	return upstream + "\n\n(NOT SUPPORTED in this Go port: " + reason + ")"
}

// toConfig converts parsed CLI flags into an executequery.Config,
// returning a typed unsupported error when the user explicitly set
// a flag whose value the Go port cannot honour.
func (rf *registeredFlags) toConfig() (executequery.Config, error) {
	cfg := executequery.Config{
		CatalogName:              rf.catalog,
		StrictNameResolutionMode: rf.strictNameResolutionMode,
	}

	// --- Supported translation ---

	for raw := range strings.SplitSeq(rf.mode, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		mode, ok := executequery.ParseMode(raw)
		if !ok {
			return executequery.Config{}, fmt.Errorf("unknown --mode value %q", raw)
		}
		cfg.Modes = append(cfg.Modes, mode)
	}
	if len(cfg.Modes) == 0 {
		cfg.Modes = []executequery.Mode{executequery.ModeAnalyze}
	}

	sm, ok := executequery.ParseSQLMode(rf.sqlMode)
	if !ok {
		return executequery.Config{}, fmt.Errorf("unknown --sql_mode %q", rf.sqlMode)
	}
	cfg.SQLMode = sm

	pm, ok := executequery.ParseProductMode(rf.productMode)
	if !ok {
		return executequery.Config{}, fmt.Errorf("unknown --product_mode %q", rf.productMode)
	}
	cfg.ProductMode = pm

	pl, ok := executequery.ParseParseLocationRecordType(rf.parseLocationRecordType)
	if !ok {
		return executequery.Config{}, fmt.Errorf("unknown --parse_location_record_type %q", rf.parseLocationRecordType)
	}
	cfg.ParseLocationRecordType = pl

	if rf.foldLiteralCast != "" {
		v := rf.foldLiteralCast == "true"
		cfg.FoldLiteralCast = &v
	}
	if rf.pruneUnusedColumns != "" {
		v := rf.pruneUnusedColumns == "true"
		cfg.PruneUnusedColumns = &v
	}

	if rf.enabledLanguageFeatures != "" {
		fs, err := executequery.ParseFeatureSet(rf.enabledLanguageFeatures)
		if err != nil {
			return executequery.Config{}, err
		}
		cfg.EnabledLanguageFeatures = fs
	}
	if rf.enabledASTRewrites != "" {
		rs, err := executequery.ParseRewriteSet(rf.enabledASTRewrites)
		if err != nil {
			return executequery.Config{}, err
		}
		cfg.EnabledASTRewrites = rs
	}
	if rf.parameters != "" {
		ps, err := executequery.ParseParameters(rf.parameters)
		if err != nil {
			return executequery.Config{}, err
		}
		cfg.Parameters = ps
	}

	// --- Unsupported sinks: populate so Validate() picks them up ---

	if rf.importPath != "" {
		cfg.ImportPaths = strings.Split(rf.importPath, ",")
	}
	if rf.tableSpec != "" {
		cfg.TableSpecs = strings.Split(rf.tableSpec, ",")
	}
	if rf.descriptorPool != "" {
		cfg.DescriptorPool = rf.descriptorPool
	}
	if rf.targetSyntax != "" {
		cfg.TargetSyntax = rf.targetSyntax
	}
	if rf.useBoxGlyphsExplicit {
		v := rf.useBoxGlyphs
		cfg.UseBoxGlyphs = &v
	}
	if rf.outputMode != "" {
		cfg.OutputMode = rf.outputMode
	}
	if rf.evaluatorMaxValueByteSizeExplicit {
		v := rf.evaluatorMaxValueByteSize
		cfg.EvaluatorMaxValueByteSize = &v
	}
	if rf.evaluatorMaxIntermediateExplicit {
		v := rf.evaluatorMaxIntermediateByteSize
		cfg.EvaluatorMaxIntermediateByteSize = &v
	}
	if rf.evaluatorScrambleExplicit {
		v := rf.evaluatorScrambleUndefinedOrderings
		cfg.EvaluatorScrambleUndefinedOrderings = &v
	}
	if rf.maxStatementsExplicit {
		v := rf.maxStatementsToExecute
		cfg.MaxStatementsToExecute = &v
	}
	if rf.web {
		cfg.Web = true
	}
	if rf.portExplicit {
		v := rf.port
		cfg.Port = &v
	}

	if err := cfg.Validate(); err != nil {
		return executequery.Config{}, err
	}
	return cfg, nil
}
