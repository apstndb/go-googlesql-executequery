// Package catalog ports the upstream `--catalog` option's selectable
// catalogs into Go. Schemas are hand-built from upstream source and
// describe.txt files (see catalog/sample.go and catalog/tpch.go);
// this Go port carries no row data, so the catalogs are usable only
// for parse / analyze. The reference evaluator that consumes data
// is not exposed by goccy/go-googlesql.
package catalog

import (
	"fmt"
	"strings"

	googlesql "github.com/goccy/go-googlesql"
)

// Name identifies a selectable catalog. The values match upstream's
// `--catalog` flag keywords.
type Name string

const (
	// None — empty catalog.
	None Name = "none"

	// Sample — the analyzer-test schema (a curated subset of
	// googlesql/testdata/sample_catalog.cc).
	Sample Name = "sample"

	// TPCH — the standard TPCH benchmark schema (8 tables, no data).
	TPCH Name = "tpch"

	// TPCHGraph — TPCH plus a property-graph view.
	// Returns an unsupported error when used (see ParseName).
	TPCHGraph Name = "tpch_graph"
)

// ParseName parses a CLI-style `--catalog` value, returning a
// non-nil error for unknown names. Returns false in the second
// return value when the parsed name is recognised but not yet
// supported by this Go port (ie tpch_graph).
func ParseName(s string) (Name, bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "none":
		return None, true, nil
	case "sample":
		return Sample, true, nil
	case "tpch":
		return TPCH, true, nil
	case "tpch_graph":
		return TPCHGraph, false, nil
	}
	return "", false, fmt.Errorf("unknown catalog %q", s)
}

// Schema is a Go-side description of a Catalog. We carry it
// alongside the *googlesql.SimpleCatalog so DESCRIBE and other
// metadata reads do not have to round-trip through the wasm
// boundary (which does not currently expose FindTable).
type Schema struct {
	Name   string
	Tables []TableSchema
}

// TableSchema describes one table.
type TableSchema struct {
	Name    string
	Columns []ColumnSchema
}

// ColumnSchema describes one column.
type ColumnSchema struct {
	Name string
	Kind googlesql.TypeKind
}

// FindTable looks up a table by name (case-insensitive). Mirrors
// upstream Catalog::FindTable's case-insensitive default.
func (s *Schema) FindTable(name string) (*TableSchema, bool) {
	if s == nil {
		return nil, false
	}
	for i := range s.Tables {
		if strings.EqualFold(s.Tables[i].Name, name) {
			return &s.Tables[i], true
		}
	}
	return nil, false
}

// Result bundles the resolved catalog (for analysis) and the Go
// schema (for DESCRIBE / introspection).
type Result struct {
	Catalog *googlesql.SimpleCatalog
	Schema  *Schema
}

// Build resolves the requested catalog and returns it ready for
// analysis. lo is required for AddBuiltinFunctionsAndTypes.
func Build(name Name, lo *googlesql.LanguageOptions, tf *googlesql.TypeFactory) (*Result, error) {
	switch name {
	case "", None:
		return buildSimple(&Schema{Name: "none"}, lo, tf)
	case Sample:
		return buildSimple(sampleSchema(), lo, tf)
	case TPCH:
		return buildSimple(tpchSchema(), lo, tf)
	case TPCHGraph:
		return nil, fmt.Errorf("catalog %q: not supported by this Go port", name)
	}
	return nil, fmt.Errorf("unknown catalog %q", name)
}

func buildSimple(schema *Schema, lo *googlesql.LanguageOptions, tf *googlesql.TypeFactory) (*Result, error) {
	cat, err := googlesql.NewSimpleCatalog(schema.Name, tf)
	if err != nil {
		return nil, fmt.Errorf("new simple catalog %q: %w", schema.Name, err)
	}
	if err := cat.AddBuiltinFunctionsAndTypes(&googlesql.BuiltinFunctionOptions{LanguageOptions: lo}); err != nil {
		return nil, fmt.Errorf("add builtins: %w", err)
	}
	for _, ts := range schema.Tables {
		tbl, err := buildTable(ts, tf)
		if err != nil {
			return nil, err
		}
		if err := cat.AddOwnedTable(tbl); err != nil {
			return nil, fmt.Errorf("add table %q: %w", ts.Name, err)
		}
	}
	return &Result{Catalog: cat, Schema: schema}, nil
}

func buildTable(ts TableSchema, tf *googlesql.TypeFactory) (*googlesql.SimpleTable, error) {
	tbl, err := googlesql.NewSimpleTable(ts.Name, -1)
	if err != nil {
		return nil, fmt.Errorf("new simple table %q: %w", ts.Name, err)
	}
	for _, c := range ts.Columns {
		typ, err := tf.MakeSimpleType(c.Kind)
		if err != nil {
			return nil, fmt.Errorf("type for %s.%s: %w", ts.Name, c.Name, err)
		}
		col, err := googlesql.NewSimpleColumn(ts.Name, c.Name, typ, false, true)
		if err != nil {
			return nil, fmt.Errorf("new column %s.%s: %w", ts.Name, c.Name, err)
		}
		if err := tbl.AddColumn(col); err != nil {
			return nil, fmt.Errorf("add column %s.%s: %w", ts.Name, c.Name, err)
		}
	}
	return tbl, nil
}

// FormatTable renders a TableSchema as plain text (one column per
// line). Used by the DESCRIBE handler.
func (t *TableSchema) Format() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Table %s:\n", t.Name)
	for _, c := range t.Columns {
		fmt.Fprintf(&b, "  %s %s\n", c.Name, typeKindName(c.Kind))
	}
	return b.String()
}

func typeKindName(k googlesql.TypeKind) string {
	// k.String() returns "TypeKindTypeFoo"; trim the prefix and
	// upper-case for readability.
	s := k.String()
	s = strings.TrimPrefix(s, "TypeKindType")
	return strings.ToUpper(s)
}
