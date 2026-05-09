package executequery

import (
	"fmt"
	"strings"

	googlesql "github.com/goccy/go-googlesql"
)

// parseTreeDebugString walks an AST rooted at root and emits a
// hierarchical pretty-print using each node's SingleNodeDebugString.
//
// goccy/go-googlesql does not expose upstream's full
// `ASTNode::DebugString()` (which recurses internally and produces
// upstream's canonical parse-tree text). Until it does, we walk the
// tree manually with NumChildren / Child and indent per depth,
// which produces semantically equivalent output without trying to
// byte-match upstream's exact format.
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
