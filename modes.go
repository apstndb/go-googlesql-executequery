package executequery

import "strings"

// Mode mirrors upstream ExecuteQueryConfig::ToolMode.
type Mode string

const (
	// ModeParse — parse the input and emit a parse-tree dump matching the C++
	// execute_query tool as closely as go-googlesql allows. Recursive AST
	// DebugString is not exported from go-googlesql (see AGENTS.md); output is
	// built manually in parseTreeDebugString using SingleNodeDebugString + tree walk.
	ModeParse Mode = "parse"

	// ModeUnparse — parse the input and emit canonical SQL via
	// go-googlesql.Unparse.
	ModeUnparse Mode = "unparse"

	// ModeAnalyze — analyze the input and emit upstream's resolved
	// DebugString. Alias accepted on the CLI: "resolve".
	ModeAnalyze Mode = "analyze"

	// ModeUnanalyze — upstream's Resolved → SQL via SQLBuilder.
	// Returns ErrUnsupportedMode wrapping ReasonModeUnanalyze.
	ModeUnanalyze Mode = "unanalyze"

	// ModeExplain — upstream's evaluator query plan.
	// Returns ErrUnsupportedMode wrapping ReasonModeExplain.
	ModeExplain Mode = "explain"

	// ModeExecute — upstream's evaluator-driven execution.
	// Returns ErrUnsupportedMode wrapping ReasonModeExecute.
	ModeExecute Mode = "execute"
)

// ParseMode normalizes a CLI-style mode string (handling the
// "resolve" / "sql_builder" aliases) and returns the corresponding
// Mode. Unknown names yield "", false.
func ParseMode(s string) (Mode, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "parse":
		return ModeParse, true
	case "unparse":
		return ModeUnparse, true
	case "analyze", "resolve":
		return ModeAnalyze, true
	case "unanalyze", "sql_builder":
		return ModeUnanalyze, true
	case "explain":
		return ModeExplain, true
	case "execute":
		return ModeExecute, true
	}
	return "", false
}

// IsSupported reports whether m can be run by this Go port.
func (m Mode) IsSupported() bool {
	switch m {
	case ModeParse, ModeUnparse, ModeAnalyze:
		return true
	}
	return false
}

// UnsupportedReason returns the per-mode reason constant for a Mode
// that is not supported. Returns "" for supported modes.
func (m Mode) UnsupportedReason() string {
	switch m {
	case ModeUnanalyze:
		return ReasonModeUnanalyze
	case ModeExplain:
		return ReasonModeExplain
	case ModeExecute:
		return ReasonModeExecute
	}
	return ""
}

// SQLMode mirrors upstream ExecuteQueryConfig::SqlMode.
type SQLMode string

const (
	// SQLModeQuery treats the input as one or more SQL statements.
	SQLModeQuery SQLMode = "query"

	// SQLModeExpression treats the input as a single SQL expression.
	SQLModeExpression SQLMode = "expression"

	// SQLModeScript treats the input as a script (multiple
	// statements with control flow). Parsing and per-statement
	// analysis are supported; script-level execution is not.
	SQLModeScript SQLMode = "script"
)

// ParseSQLMode parses a CLI-style sql_mode value.
func ParseSQLMode(s string) (SQLMode, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "query":
		return SQLModeQuery, true
	case "expression":
		return SQLModeExpression, true
	case "script":
		return SQLModeScript, true
	}
	return "", false
}

// ProductMode mirrors googlesql::ProductMode.
type ProductMode string

const (
	// ProductModeInternal — upstream's PRODUCT_INTERNAL (e.g.
	// supports proto types and DOUBLE).
	ProductModeInternal ProductMode = "internal"

	// ProductModeExternal — upstream's PRODUCT_EXTERNAL.
	ProductModeExternal ProductMode = "external"
)

// ParseProductMode parses a CLI-style product_mode value.
func ParseProductMode(s string) (ProductMode, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "internal":
		return ProductModeInternal, true
	case "external":
		return ProductModeExternal, true
	}
	return "", false
}

// ParseLocationRecordType mirrors
// googlesql::ParseLocationRecordType.
type ParseLocationRecordType string

const (
	// ParseLocationRecordNone — record no parse locations.
	ParseLocationRecordNone ParseLocationRecordType = "NONE"

	// ParseLocationRecordFullNodeScope — record full-node-scope
	// locations.
	ParseLocationRecordFullNodeScope ParseLocationRecordType = "FULL_NODE_SCOPE"

	// ParseLocationRecordCodeSearch — record locations suitable for
	// code-search.
	ParseLocationRecordCodeSearch ParseLocationRecordType = "CODE_SEARCH"
)

// ParseParseLocationRecordType parses a CLI-style value. Names match
// upstream's enum spelling (case-insensitive).
func ParseParseLocationRecordType(s string) (ParseLocationRecordType, bool) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "", "NONE":
		return ParseLocationRecordNone, true
	case "FULL_NODE_SCOPE":
		return ParseLocationRecordFullNodeScope, true
	case "CODE_SEARCH":
		return ParseLocationRecordCodeSearch, true
	}
	return "", false
}
