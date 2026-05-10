package executequery

import (
	"fmt"
	"strings"

	googlesql "github.com/goccy/go-googlesql"
)

// parseTreeDebugString walks an AST rooted at root and emits a hierarchical
// parse-tree dump shaped like the C++ execute_query reference binary.
//
// Same constraint family as AGENTS.md "Modes": go-googlesql does not export
// the recursive AST DebugString that upstream uses for --mode=parse output—only
// SingleNodeDebugString (per-node, single-line). Unblocked when go-googlesql
// exposes a recursive ASTNode.DebugString (or equivalent).
//
// Until then we walk NumChildren / Child ourselves and append byte spans from
// Location() (withParseLocationSuffix), because SingleNodeDebugString alone does
// not include source ranges and does not recurse.
func parseTreeDebugString(root googlesql.ASTNode) (string, error) {
	var b strings.Builder
	if err := walkPrintAST(&b, root, 0); err != nil {
		return "", err
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func walkPrintAST(b *strings.Builder, n googlesql.ASTNode, depth int) error {
	if n == nil {
		return nil
	}
	line, err := n.SingleNodeDebugString()
	if err != nil {
		return fmt.Errorf("debug string: %w", err)
	}
	line = withParseLocationSuffix(n, line)
	for range depth {
		b.WriteString("  ")
	}
	b.WriteString(line)
	b.WriteByte('\n')
	num, err := n.NumChildren()
	if err != nil {
		return fmt.Errorf("num children: %w", err)
	}
	for i := range num {
		c, err := n.Child(i)
		if err != nil {
			return fmt.Errorf("child %d: %w", i, err)
		}
		if c == nil {
			continue
		}
		if err := walkPrintAST(b, c, depth+1); err != nil {
			return err
		}
	}
	return nil
}

// withParseLocationSuffix appends ` [start-end]` from ASTNode.Location when
// valid. This is local formatting work: go-googlesql does not combine spans into
// the recursive DebugString line format—that API is not exported (see above).
func withParseLocationSuffix(n googlesql.ASTNode, line string) string {
	if n == nil {
		return line
	}
	loc, err := n.Location()
	if err != nil || loc == nil {
		return line
	}
	ok, err := loc.IsValid()
	if err != nil || !ok {
		return line
	}
	start, err := loc.Start()
	if err != nil || start == nil {
		return line
	}
	end, err := loc.End()
	if err != nil || end == nil {
		return line
	}
	sOff, err := start.GetByteOffset()
	if err != nil {
		return line
	}
	eOff, err := end.GetByteOffset()
	if err != nil {
		return line
	}
	return fmt.Sprintf("%s [%d-%d]", line, sOff, eOff)
}
