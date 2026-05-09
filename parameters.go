package executequery

import (
	"fmt"
	"strconv"
	"strings"

	googlesql "github.com/goccy/go-googlesql"
)

// QueryParameter is one entry registered with the analyzer for an
// `@name` parameter reference. The Type field is what
// AnalyzerOptions.AddQueryParameter actually consumes; Literal is
// preserved for diagnostics and would be the Value for execute mode
// (unsupported in this Go port).
type QueryParameter struct {
	Name    string
	Type    googlesql.TypeKind
	Literal string // raw text from the CLI flag
}

// ParseParameters parses upstream's `--parameters` flag format,
// `name=literal,name=literal,...`.
//
// Implementation note: upstream
// `googlesql::AnalyzeExpression`
// (third_party/googlesql/googlesql/public/analyzer.h:127) is the
// general path for inferring a parameter literal's type, and
// go-googlesql does expose it — but routing parameter parsing
// through the analyzer would require constructing an AnalyzerOptions
// + Catalog + TypeFactory just to type a `42` or a `'foo'`. Until we
// have a real need for richer literal forms (array / struct /
// NUMERIC / etc.), the simple regex-style recognition below is
// cheaper and easier to read; switch to AnalyzeExpression when the
// trade-off flips.
//
//	42        → INT64
//	3.14      → DOUBLE
//	'foo' or "foo"  → STRING
//	TRUE / FALSE     → BOOL
//	NULL             → error (no type information)
func ParseParameters(s string) ([]QueryParameter, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var params []QueryParameter
	for _, part := range splitTopLevel(s, ',') {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		namePart, literalPart, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("parameters: %q: expected name=value", part)
		}
		name := strings.TrimSpace(namePart)
		literal := strings.TrimSpace(literalPart)
		if name == "" {
			return nil, fmt.Errorf("parameters: %q: empty name", part)
		}
		kind, err := inferTypeKind(literal)
		if err != nil {
			return nil, fmt.Errorf("parameters: %q: %w", part, err)
		}
		params = append(params, QueryParameter{Name: name, Type: kind, Literal: literal})
	}
	return params, nil
}

func inferTypeKind(literal string) (googlesql.TypeKind, error) {
	switch strings.ToUpper(literal) {
	case "":
		return 0, fmt.Errorf("empty literal")
	case "TRUE", "FALSE":
		return googlesql.TypeKindTypeBool, nil
	case "NULL":
		return 0, fmt.Errorf("NULL literals carry no type information; provide a typed literal")
	}
	if (strings.HasPrefix(literal, "'") && strings.HasSuffix(literal, "'")) ||
		(strings.HasPrefix(literal, `"`) && strings.HasSuffix(literal, `"`)) {
		if len(literal) < 2 {
			return 0, fmt.Errorf("malformed string literal")
		}
		return googlesql.TypeKindTypeString, nil
	}
	if _, err := strconv.ParseInt(literal, 10, 64); err == nil {
		return googlesql.TypeKindTypeInt64, nil
	}
	if _, err := strconv.ParseFloat(literal, 64); err == nil {
		return googlesql.TypeKindTypeDouble, nil
	}
	return 0, fmt.Errorf("unrecognised literal %q (supported: integers, floats, 'string', \"string\", TRUE, FALSE)", literal)
}

// splitTopLevel splits s on sep characters that are not inside
// matched single- or double-quotes. Backslash-escapes inside quotes
// are honoured.
func splitTopLevel(s string, sep byte) []string {
	var out []string
	var b strings.Builder
	inSingle, inDouble, escape := false, false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case escape:
			b.WriteByte(c)
			escape = false
		case c == '\\' && (inSingle || inDouble):
			b.WriteByte(c)
			escape = true
		case c == '\'' && !inDouble:
			inSingle = !inSingle
			b.WriteByte(c)
		case c == '"' && !inSingle:
			inDouble = !inDouble
			b.WriteByte(c)
		case c == sep && !inSingle && !inDouble:
			out = append(out, b.String())
			b.Reset()
		default:
			b.WriteByte(c)
		}
	}
	out = append(out, b.String())
	return out
}
