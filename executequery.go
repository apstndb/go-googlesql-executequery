package executequery

import (
	"context"
	"fmt"

	googlesql "github.com/goccy/go-googlesql"

	"github.com/apstndb/go-googlesql-executequery/catalog"
)

// Run is the top-level entry point. It validates cfg, constructs
// the GoogleSQL parser/analyzer state, and dispatches sql through
// each requested Mode for each statement, emitting via w.
//
// Run does not initialise go-googlesql; callers must call
// cache.Setup or googlesql.Init once before the first call.
func Run(ctx context.Context, sql string, cfg Config, w Writer) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if w == nil {
		return fmt.Errorf("nil writer")
	}

	tf, err := googlesql.NewTypeFactory()
	if err != nil {
		return fmt.Errorf("new type factory: %w", err)
	}

	lo, err := cfg.buildLanguageOptions()
	if err != nil {
		return err
	}

	cat, err := selectCatalog(cfg.CatalogName, lo, tf)
	if err != nil {
		return err
	}

	ao, err := cfg.buildAnalyzerOptions(lo, tf)
	if err != nil {
		return err
	}

	po, err := googlesql.NewParserOptions()
	if err != nil {
		return fmt.Errorf("new parser options: %w", err)
	}
	// Workaround [go-googlesql v0.2.1]: ParserOptions.SetLanguageOptions silently *moves-from* its
	// argument on the wasm side, leaving the caller's
	// *LanguageOptions handle pointing at a default-constructed
	// instance (no language features enabled).
	//
	// Upstream C++ API: googlesql::ParserOptions::set_language_options
	// (LanguageOptions language_options) at
	// third_party/googlesql/googlesql/parser/parser.h:123 — note
	// upstream takes the argument by value, so a Go binding that
	// faithfully mirrors C++ would copy under the hood.
	//
	// Natural Go code:
	//   po.SetLanguageOptions(lo)   // share the same LO with the analyzer
	//
	// Instead, we build a dedicated parserLO so the analyzer's LO
	// survives unchanged. Unblocked when go-googlesql either copies the
	// argument (matching the C++ by-value contract) or documents the
	// move-from semantics so callers can opt in.
	parserLO, err := cfg.buildLanguageOptions()
	if err != nil {
		return err
	}
	if err := po.SetLanguageOptions(parserLO); err != nil {
		return fmt.Errorf("set parser language options: %w", err)
	}

	state := &runState{
		ctx: ctx, cfg: cfg, w: w,
		tf: tf, lo: lo, ao: ao, po: po, cat: cat,
		sql: sql,
	}

	switch cfg.effectiveSQLMode() {
	case SQLModeQuery:
		return state.runQueryMode()
	case SQLModeExpression:
		return state.runExpressionMode()
	case SQLModeScript:
		return state.runScriptMode()
	}
	return sqlModeUnsupportedf(string(cfg.effectiveSQLMode()), "unknown sql_mode")
}

func selectCatalog(name string, lo *googlesql.LanguageOptions, tf *googlesql.TypeFactory) (*catalog.Result, error) {
	parsed, supported, err := catalog.ParseName(name)
	if err != nil {
		return nil, err
	}
	if !supported {
		// Reserved for future catalogs that are recognised but cannot
		// be built; today every recognised name is also supported.
		return nil, catalogUnsupportedf(string(parsed), "catalog recognised but not yet built in this Go port")
	}
	return catalog.Build(parsed, lo, tf)
}

type runState struct {
	ctx context.Context
	cfg Config
	w   Writer
	sql string

	tf  *googlesql.TypeFactory
	lo  *googlesql.LanguageOptions
	ao  *googlesql.AnalyzerOptions
	po  *googlesql.ParserOptions
	cat *catalog.Result
}

// runQueryMode iterates statements via ParseNextStatement and runs
// each requested Mode for each one.
func (s *runState) runQueryMode() error {
	loc, err := googlesql.NewParseResumeLocationFromString(s.sql)
	if err != nil {
		return fmt.Errorf("new parse resume location: %w", err)
	}
	first := true
	for {
		atStart, err := loc.BytePosition()
		if err != nil {
			return fmt.Errorf("byte position: %w", err)
		}
		if int(atStart) >= len(s.sql) {
			return nil
		}
		if err := s.w.StartStatement(first); err != nil {
			return err
		}
		first = false
		stmtErr := s.processOneStatement(loc)
		errMsg := ""
		if stmtErr != nil {
			errMsg = stmtErr.Error()
		}
		if err := s.w.FlushStatement(false, errMsg); err != nil {
			return err
		}
		if stmtErr != nil {
			return stmtErr
		}
	}
}

// processOneStatement parses one statement starting at loc, runs each
// requested Mode, and advances loc.
func (s *runState) processOneStatement(loc *googlesql.ParseResumeLocation) error {
	out, err := googlesql.ParseNextStatement(loc, s.po)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	stmt, err := out.Statement()
	if err != nil {
		return fmt.Errorf("get statement: %w", err)
	}

	if handled, err := describeIfApplicable(stmt, s.cat.Schema, s.w); handled {
		return err
	}

	for _, mode := range s.cfg.effectiveModes() {
		if err := s.emitMode(mode, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *runState) emitMode(mode Mode, stmt googlesql.ASTStatementNode) error {
	switch mode {
	case ModeParse:
		text, err := parseTreeDebugString(stmt)
		if err != nil {
			return fmt.Errorf("parse mode: %w", err)
		}
		return s.w.Parsed(text)
	case ModeUnparse:
		text, err := googlesql.Unparse(stmt)
		if err != nil {
			return fmt.Errorf("unparse: %w", err)
		}
		return s.w.Unparsed(text)
	case ModeAnalyze:
		ao, err := s.cfg.buildAnalyzerOptions(s.lo, s.tf)
		if err != nil {
			return err
		}
		// AnalyzeStatementFromParserAST takes the parsed AST and
		// avoids a second parse pass.
		out, err := googlesql.AnalyzeStatementFromParserAST(stmt, ao, s.sql, s.cat.Catalog, s.tf)
		if err != nil {
			return fmt.Errorf("analyze: %w", err)
		}
		resolved, err := out.ResolvedStatement()
		if err != nil {
			return fmt.Errorf("get resolved: %w", err)
		}
		text, err := resolved.DebugString()
		if err != nil {
			return fmt.Errorf("resolved debug: %w", err)
		}
		return s.w.Resolved(text)
	}
	return modeUnsupportedf(string(mode), mode.UnsupportedReason())
}

// runExpressionMode parses the entire input as one expression and
// runs the requested modes.
func (s *runState) runExpressionMode() error {
	if err := s.w.StartStatement(true); err != nil {
		return err
	}
	out, err := googlesql.ParseExpression(s.sql, s.po)
	if err != nil {
		return s.w.FlushStatement(true, err.Error())
	}
	expr, err := out.Expression()
	if err != nil {
		return s.w.FlushStatement(true, err.Error())
	}
	for _, mode := range s.cfg.effectiveModes() {
		switch mode {
		case ModeParse:
			text, err := parseTreeDebugString(expr)
			if err != nil {
				return s.w.FlushStatement(true, err.Error())
			}
			if err := s.w.Parsed(text); err != nil {
				return err
			}
		case ModeUnparse:
			text, err := googlesql.Unparse(expr)
			if err != nil {
				return s.w.FlushStatement(true, err.Error())
			}
			if err := s.w.Unparsed(text); err != nil {
				return err
			}
		case ModeAnalyze:
			analOut, err := googlesql.AnalyzeExpression(s.sql, s.ao, s.cat.Catalog, s.tf)
			if err != nil {
				return s.w.FlushStatement(true, err.Error())
			}
			resolved, err := analOut.ResolvedExpr()
			if err != nil {
				return s.w.FlushStatement(true, err.Error())
			}
			text, err := resolved.DebugString()
			if err != nil {
				return s.w.FlushStatement(true, err.Error())
			}
			if err := s.w.Resolved(text); err != nil {
				return err
			}
		default:
			return modeUnsupportedf(string(mode), mode.UnsupportedReason())
		}
	}
	return s.w.FlushStatement(true, "")
}

// runScriptMode parses the input as a script and runs the requested
// modes per top-level statement.
func (s *runState) runScriptMode() error {
	out, err := googlesql.ParseScript(s.sql, s.po, nil)
	if err != nil {
		return fmt.Errorf("parse script: %w", err)
	}
	script, err := out.Node()
	if err != nil {
		return fmt.Errorf("get script node: %w", err)
	}
	if err := s.w.StartStatement(true); err != nil {
		return err
	}
	for _, mode := range s.cfg.effectiveModes() {
		switch mode {
		case ModeParse:
			text, err := parseTreeDebugString(script)
			if err != nil {
				return s.w.FlushStatement(true, err.Error())
			}
			if err := s.w.Parsed(text); err != nil {
				return err
			}
		case ModeUnparse:
			text, err := googlesql.Unparse(script)
			if err != nil {
				return s.w.FlushStatement(true, err.Error())
			}
			if err := s.w.Unparsed(text); err != nil {
				return err
			}
		case ModeAnalyze:
			// Not yet implemented: per-statement analyze on a parsed
			// script. There is no `AnalyzeScript` symbol upstream
			// (third_party/googlesql/googlesql/public/analyzer.h has
			// AnalyzeStatement / AnalyzeStatementFromParserAST /
			// AnalyzeNextStatement / AnalyzeExpression but no
			// script-level entry); upstream callers iterate the
			// `ASTScript`'s `StatementListNode().GetChildren()` and
			// call `AnalyzeStatementFromParserAST` per statement.
			// `go-googlesql` exposes the same primitives
			// (`ParserOutput.Script`, `ASTScript.StatementListNode`,
			// `AnalyzeStatementFromParserAST`), so this is a missing
			// implementation rather than a binding gap.
			if err := s.w.Resolved("(analyze in script mode is not yet implemented in this Go port; emit per-statement parse / analyze instead)"); err != nil {
				return err
			}
		default:
			return modeUnsupportedf(string(mode), mode.UnsupportedReason())
		}
	}
	return s.w.FlushStatement(true, "")
}
