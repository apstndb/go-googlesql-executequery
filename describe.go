package executequery

import (
	"fmt"
	"strings"

	googlesql "github.com/goccy/go-googlesql"

	"github.com/apstndb/go-googlesql-executequery/catalog"
)

// describeIfApplicable inspects stmt and, if it is a DESCRIBE
// statement, formats the description into Writer.Described.
//
// Returns (true, nil) on a successful DESCRIBE emit, (false, nil)
// when stmt is not a DESCRIBE, and (true, err) when stmt is a
// DESCRIBE but the lookup or emit failed.
//
// Mirrors the special-case in upstream's execute_query_tool.cc which
// pre-empts analysis for DESCRIBE so the lookup goes against the
// active catalog.
func describeIfApplicable(stmt googlesql.ASTStatementNode, schema *catalog.Schema, w Writer) (bool, error) {
	d, ok := stmt.(*googlesql.ASTDescribeStatement)
	if !ok {
		return false, nil
	}
	pathExpr, err := d.Name()
	if err != nil {
		return true, fmt.Errorf("describe: read path expression: %w", err)
	}
	parts, err := pathExpr.ToIdentifierVector()
	if err != nil {
		return true, fmt.Errorf("describe: read identifiers: %w", err)
	}
	name := strings.Join(parts, ".")
	if t, ok := schema.FindTable(name); ok {
		// Inner text matches upstream ExecuteDescribe table formatting
		// (execute_query_tool.cc); box matches default box output (output_query_result.cc).
		return true, w.Described(boxUnicodeSingleColumn("Describe", t.Format()))
	}
	return true, w.Described(fmt.Sprintf("Object %q not found in catalog %q", name, schema.Name))
}
