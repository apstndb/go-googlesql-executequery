package executequery

import (
	"fmt"
	"io"
)

// Writer is the sink that Run streams its output through.
//
// The interface is a deliberately reduced port of upstream's
// ExecuteQueryWriter. Only the methods covering supported modes
// (parse / unparse / analyze / describe) are present. Each method
// returns an error so an io.Writer-backed implementation can
// propagate write failures.
type Writer interface {
	// StatementText writes the source SQL of the current statement
	// (called once per statement before any other emit method).
	StatementText(text string) error

	// Parsed emits parse-mode output (the parse tree).
	Parsed(debug string) error

	// Unparsed emits unparse-mode output (canonical SQL).
	Unparsed(sql string) error

	// Resolved emits analyze-mode output (the resolved AST).
	Resolved(debug string) error

	// Described emits DESCRIBE-statement output (table schema /
	// function signature / etc.).
	Described(text string) error

	// StartStatement is called at the start of each statement.
	// isFirst is true for the first statement only.
	StartStatement(isFirst bool) error

	// FlushStatement is called at the end of each statement, or
	// when an error truncates it. atEnd is true when the entire
	// input has been consumed; errMsg is non-empty iff the
	// statement failed.
	FlushStatement(atEnd bool, errMsg string) error
}

// NewTextWriter returns a Writer that emits plain-text CLI output to w.
//
// Parse and unparse modes stream output in the same layout as the upstream
// C++ execute_query tool (parse tree with byte-offset spans, then a blank
// line, then canonical SQL) rather than labeled `[parse]` / `[unparse]` sections.
//
// Analyze and describe output remain labeled (`[analyze]`, `[describe]`) for readability.
func NewTextWriter(w io.Writer) Writer { return &textWriter{w: w} }

type textWriter struct {
	w io.Writer
	// first is true after emitting any section in the current statement so the
	// next labeled section (analyze/describe) is prefixed with a newline.
	first bool
	// parseEmitted is true after Parsed() until Unparsed() consumes the blank-line gap.
	parseEmitted bool
}

func (t *textWriter) writeSection(label, body string) error {
	body = trimTrailingNewlines(body)
	if t.first {
		if _, err := io.WriteString(t.w, "\n"); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(t.w, "[%s]\n%s\n", label, indent(body, "  ")); err != nil {
		return err
	}
	t.first = true
	return nil
}

func (t *textWriter) StatementText(text string) error { return t.writeSection("statement", text) }
func (t *textWriter) Parsed(debug string) error {
	debug = trimTrailingNewlines(debug)
	if _, err := fmt.Fprintf(t.w, "%s\n", debug); err != nil {
		return err
	}
	t.first = true
	t.parseEmitted = true
	return nil
}

func (t *textWriter) Unparsed(sql string) error {
	sql = trimTrailingNewlines(sql)
	prefix := ""
	if t.parseEmitted {
		prefix = "\n"
		t.parseEmitted = false
	}
	if _, err := fmt.Fprintf(t.w, "%s%s\n", prefix, sql); err != nil {
		return err
	}
	t.first = true
	return nil
}
func (t *textWriter) Resolved(debug string) error { return t.writeSection("analyze", debug) }
func (t *textWriter) Described(text string) error { return t.writeSection("describe", text) }

func (t *textWriter) StartStatement(isFirst bool) error {
	if !isFirst {
		if _, err := io.WriteString(t.w, "\n----\n"); err != nil {
			return err
		}
	}
	t.first = false
	t.parseEmitted = false
	return nil
}

func (t *textWriter) FlushStatement(_ bool, errMsg string) error {
	if errMsg != "" {
		if _, err := fmt.Fprintf(t.w, "\n[error]\n  %s\n", errMsg); err != nil {
			return err
		}
	}
	return nil
}

func trimTrailingNewlines(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

func indent(s, prefix string) string {
	if s == "" {
		return prefix
	}
	out := make([]byte, 0, len(s)+len(prefix))
	atStart := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if atStart {
			out = append(out, prefix...)
			atStart = false
		}
		out = append(out, c)
		if c == '\n' {
			atStart = true
		}
	}
	return string(out)
}
