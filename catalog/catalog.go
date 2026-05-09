// Package catalog ports the upstream `--catalog` option's selectable
// catalogs into Go. Schemas are hand-built from upstream source and
// describe.txt files (see catalog/sample.go and catalog/tpch.go);
// this Go port carries no row data, so the catalogs are usable only
// for parse / analyze. The reference evaluator that consumes data
// is not exposed by go-googlesql.
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
		return TPCHGraph, true, nil
	}
	return "", false, fmt.Errorf("unknown catalog %q", s)
}

// Schema is a Go-side description of a Catalog. We carry it
// alongside the *googlesql.SimpleCatalog so DESCRIBE and other
// metadata reads do not have to round-trip through the wasm
// boundary.
//
// Workaround [go-googlesql v0.2.1]: `SimpleCatalog.FindTable` (and equivalent table-list accessors that
// hand back a usable handle) are not exposed, so once a SimpleTable
// has been registered via `AddOwnedTable` we cannot read its columns
// back through the catalog.
//
// Upstream C++ API:
//   - googlesql::Catalog::FindTable(absl::Span<const std::string>,
//     const Table**)
//   - googlesql::Table::GetColumn(int) / NumColumns()
//   - googlesql::Column::GetType() / Name()
//
// (third_party/googlesql/googlesql/public/simple_catalog.h:76,
// the `Catalog` base in catalog.h.)
//
// Natural Go code:
//
//	tbl, _ := cat.FindTable("Foo")
//	for i := int32(0); i < tbl.NumColumns(); i++ {
//	    col, _ := tbl.GetColumn(i)
//	    name, _ := col.Name()
//	    typ,  _ := col.GetType()
//	    ...
//	}
//
// Instead, we keep an immutable Go-side mirror of the schema.
// Unblocked when go-googlesql exports `SimpleCatalog.FindTable` (and the
// `Column`/`Type` getters needed to render upstream-format output).
type Schema struct {
	Name   string
	Tables []TableSchema
}

// TableSchema describes one table.
type TableSchema struct {
	Name string
	// PrimaryKey lists column names that form the primary key, in
	// order. Empty when the table has no declared primary key.
	PrimaryKey []string
	Columns    []ColumnSchema
}

// ColumnSchema describes one column.
type ColumnSchema struct {
	Name string
	Kind googlesql.TypeKind
	// TypeName overrides the type label rendered by Format() (and is
	// used for non-scalar types like PROTO<...> / ENUM<...> that the
	// Kind enum cannot describe alone). Empty for simple types.
	TypeName string
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
		return buildSimple(&Schema{Name: "none"}, lo, tf, nil)
	case Sample:
		return buildSimple(sampleSchema(), lo, tf, sampleProtoPostBuild)
	case TPCH:
		return buildSimple(tpchSchema(), lo, tf, nil)
	case TPCHGraph:
		return buildTPCHGraph(lo, tf)
	}
	return nil, fmt.Errorf("unknown catalog %q", name)
}

// PostBuild is the per-catalog hook invoked after schema.Tables are
// constructed but before they are handed to the SimpleCatalog. It can
// mutate the freshly-built SimpleTables (e.g. add pseudo-columns,
// see tpch_graph) and/or attach extras to the catalog directly
// (e.g. SetDescriptorPool, AddType, AddOwnedTable for proto-typed
// tables that don't fit the Schema → buildTable flow).
//
// schema is the catalog's Schema; the hook may append to schema.Tables
// to surface the extra tables in DESCRIBE output. tables maps
// lower-cased schema-table name to its handle. After buildSimple
// returns from the hook, AddOwnedTable is called for every schema
// table — extras the hook AddOwnedTabled itself are *not* re-added.
//
// Workaround [go-googlesql v0.2.1]: SimpleCatalog.AddOwnedTable calls clearPtrAny(table) on its argument
// after the wasm round-trip, leaving any retained *SimpleTable handle
// null. The pre-AddOwnedTable hook is the only safe place to mutate
// tables.
//
// Upstream C++ API:
// googlesql::SimpleCatalog::AddOwnedTable
// (third_party/googlesql/googlesql/public/simple_catalog.h:184-192) —
// upstream takes ownership of the unique_ptr but does not invalidate
// any raw `Table*` aliases the caller may hold; live mutation through
// such an alias is legal in C++.
//
// Natural Go code:
//
//	cat.AddOwnedTable(tbl)
//	tbl.AddColumn(...)              // or, equivalently:
//	live, _ := cat.FindTable("Foo")  // then live.AddColumn(...)
//
// Instead, callers must populate tables fully before AddOwnedTable, or
// AddOwnedTable in their own hook. Unblocked when go-googlesql either
// stops clearing the handle or exposes `SimpleCatalog.FindTable` so we
// can retrieve a live handle after registration.
type PostBuild func(cat *googlesql.SimpleCatalog, schema *Schema, tables map[string]*googlesql.SimpleTable, tf *googlesql.TypeFactory) error

// buildSimple registers all tables from schema in a fresh
// SimpleCatalog. The optional postBuild hook runs after the schema
// tables are built but before AddOwnedTable, so callers can attach
// extra columns to live SimpleTable handles, register a
// DescriptorPool, or AddOwnedTable extra tables of their own.
func buildSimple(schema *Schema, lo *googlesql.LanguageOptions, tf *googlesql.TypeFactory, postBuild PostBuild) (*Result, error) {
	cat, err := googlesql.NewSimpleCatalog(schema.Name, tf)
	if err != nil {
		return nil, fmt.Errorf("new simple catalog %q: %w", schema.Name, err)
	}
	if err := cat.AddBuiltinFunctionsAndTypes(&googlesql.BuiltinFunctionOptions{LanguageOptions: lo}); err != nil {
		return nil, fmt.Errorf("add builtins: %w", err)
	}
	// Snapshot the schema-table list before the hook runs so that
	// hooks which mirror catalog-only tables back into schema (e.g.
	// the sample catalog's proto/enum tables, registered directly via
	// cat.AddOwnedTable inside the hook) don't make this loop try to
	// re-add them.
	originalTables := schema.Tables
	tables := make(map[string]*googlesql.SimpleTable, len(originalTables))
	for _, ts := range originalTables {
		tbl, err := buildTable(ts, tf)
		if err != nil {
			return nil, err
		}
		tables[strings.ToLower(ts.Name)] = tbl
	}
	if postBuild != nil {
		if err := postBuild(cat, schema, tables, tf); err != nil {
			return nil, err
		}
	}
	for _, ts := range originalTables {
		tbl := tables[strings.ToLower(ts.Name)]
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
	colIndex := make(map[string]int32, len(ts.Columns))
	for i, c := range ts.Columns {
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
		colIndex[strings.ToLower(c.Name)] = int32(i)
	}
	if len(ts.PrimaryKey) > 0 {
		idx := make([]int32, len(ts.PrimaryKey))
		for i, name := range ts.PrimaryKey {
			j, ok := colIndex[strings.ToLower(name)]
			if !ok {
				return nil, fmt.Errorf("primary key for %s: column %q not in table", ts.Name, name)
			}
			idx[i] = j
		}
		if err := tbl.SetPrimaryKey(idx); err != nil {
			return nil, fmt.Errorf("set primary key for %s: %w", ts.Name, err)
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
		label := c.TypeName
		if label == "" {
			label = typeKindName(c.Kind)
		}
		fmt.Fprintf(&b, "  %s %s\n", c.Name, label)
	}
	if len(t.PrimaryKey) > 0 {
		fmt.Fprintf(&b, "Primary key: (%s)\n", strings.Join(t.PrimaryKey, ", "))
	}
	return b.String()
}

// typeKindName renders a TypeKind as its upstream-style upper-case
// name (e.g. INT64).
//
// Workaround [go-googlesql v0.2.1]: there is no Go-side accessor
// for the user-facing form of a TypeKind; only the Go enum's
// `String()` method is exported, which returns "TypeKindTypeInt64".
//
// Upstream C++ API: googlesql::Type::TypeKindToString(TypeKind,
// ProductMode, bool use_external_float32)
// (third_party/googlesql/googlesql/public/types/type.h:572) —
// returns the user-facing name ("INT64").
//
// Natural Go code:
//
//	name, _ := googlesql.TypeKindToString(k, googlesql.ProductModeProductInternal)
//
// Instead, strip the "TypeKindType" prefix from the Go enum's
// `String()` output and upper-case the remainder. Unblocked when
// go-googlesql exports `TypeKindToString` (or an equivalent
// short-name accessor on TypeKind).
func typeKindName(k googlesql.TypeKind) string {
	s := k.String()
	s = strings.TrimPrefix(s, "TypeKindType")
	return strings.ToUpper(s)
}
