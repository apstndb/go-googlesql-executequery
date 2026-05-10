package executequery

import (
	"encoding/json"
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
func describeIfApplicable(stmt googlesql.ASTStatementNode, schema *catalog.Schema, w Writer, cfg *Config) (bool, error) {
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
	outMode := strings.ToLower(strings.TrimSpace(cfg.OutputMode))

	emit := func(plain string) error {
		switch outMode {
		case "json":
			s, err := describeJSONRows(plain)
			if err != nil {
				return err
			}
			return w.Described(s)
		default:
			// "" or "box": Unicode box like upstream default output_mode=box.
			return w.Described(boxUnicodeSingleColumn("Describe", plain))
		}
	}

	if t, ok := schema.FindTable(name); ok {
		return true, emit(t.Format())
	}
	msg := fmt.Sprintf("Object %q not found in catalog %q", name, schema.Name)
	return true, emit(msg)
}

// describeJSONRows wraps plain DESCRIBE text like upstream execute_query
// --output_mode=json (one row, column "Describe").
func describeJSONRows(describePlain string) (string, error) {
	type rowObj struct {
		Describe string `json:"Describe"`
	}
	v := struct {
		Row []rowObj `json:"row"`
	}{
		Row: []rowObj{{Describe: describePlain}},
	}
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return "", fmt.Errorf("describe json: %w", err)
	}
	return string(b), nil
}
