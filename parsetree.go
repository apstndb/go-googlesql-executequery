package executequery

import (
	"fmt"
	"strings"

	googlesql "github.com/goccy/go-googlesql"
)

// parseTreeDebugString walks an AST rooted at root and emits a
// hierarchical pretty-print using each node's SingleNodeDebugString.
//
// Workaround for go-googlesql v0.2.1: the recursive multi-line
// debug formatter upstream uses for `--mode=parse` output is not
// exposed; only the per-node single-line formatter is.
//
// Upstream C++ API: googlesql::ASTNode::DebugString(int max_depth)
// (third_party/googlesql/googlesql/parser/ast_node.h:243).
// `SingleNodeDebugString` (line 92) IS bound by go-googlesql; the
// recursive variant is not.
//
// Natural Go code:
//
//	text, err := root.DebugString()
//
// Instead, we walk the tree manually with NumChildren / Child and
// indent per depth, which produces semantically equivalent output
// without trying to byte-match upstream's exact format. Unblocked
// when go-googlesql exports `ASTNode.DebugString`.
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
