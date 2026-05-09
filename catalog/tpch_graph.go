package catalog

import (
	"fmt"
	"strings"

	googlesql "github.com/goccy/go-googlesql"
)

// buildTPCHGraph builds the tpch catalog and adds the join
// pseudo-columns described by upstream's `--catalog=tpch_graph` mode.
//
// Ported from
// third_party/googlesql/googlesql/examples/tpch/catalog/tpch_catalog.cc
// (AddJoinColumns / AddJoinColumn / AddOneJoinColumn).
//
// Workaround for go-googlesql v0.2.1: upstream attaches a
// Column::JoinColumnAttributes to each pseudo-column so the analyzer
// can reason about the join. go-googlesql exports the
// OptionalJoinColumnAttributes handle type but not its constructor or
// any way to thread it through SimpleColumn construction.
//
// Upstream C++ API: googlesql::SimpleColumn::Attributes
// (third_party/googlesql/googlesql/public/simple_catalog.h:1180-1189)
// is a struct passed *at construction* via the
// `SimpleColumn(absl::string_view table_full_name, absl::string_view
// name, const Type* type, Attributes attributes)` overload
// (simple_catalog.h:1198-1214). The attribute fields are
// `is_pseudo_column`, `is_writable_column`, and
// `std::optional<JoinColumnAttributes> join_column`. The
// construction-time pattern is what tpch_catalog.cc:384-396 uses.
//
// Natural Go code (a constructor variant accepting Attributes):
//
//	col, _ := googlesql.NewSimpleColumnWithAttributes(
//	    tableFullName, columnName, rowType,
//	    googlesql.SimpleColumnAttributes{
//	        IsPseudoColumn:   true,
//	        IsWritableColumn: false,
//	        JoinColumn: googlesql.NewJoinColumnAttributes(
//	            boundColumns, sourceTable, sourceColumns, isMultiRow),
//	    })
//
// Instead, the pseudo-columns we register carry no
// JoinColumnAttributes. Walking the pseudo-columns
// (`Customer.Orders`) still resolves through the type system via the
// RowType, but operations that depend on JoinColumnAttributes (e.g.
// upstream's join-flattening rewrite) will not behave identically.
// Unblocked when go-googlesql exposes the `Attributes`-taking
// `SimpleColumn` constructor (and binds the
// `JoinColumnAttributes` struct).
func buildTPCHGraph(lo *googlesql.LanguageOptions, tf *googlesql.TypeFactory) (*Result, error) {
	schema := tpchSchema()
	schema.Name = "tpch_graph"

	// (table1, table2, columns1, columns2, isMulti1, isMulti2). isMulti
	// is true on the side that holds *many* rows for one row of the
	// other side, mirroring the C++ `is_multi*` arguments.
	type joinSpec struct {
		t1, t2     string
		c1, c2     []string
		multi1, m2 bool
	}
	joins := []joinSpec{
		{"Customer", "Orders", []string{"C_CUSTKEY"}, []string{"O_CUSTKEY"}, false, true},
		{"Orders", "LineItem", []string{"O_ORDERKEY"}, []string{"L_ORDERKEY"}, false, true},
		{"Region", "Nation", []string{"R_REGIONKEY"}, []string{"N_REGIONKEY"}, false, true},
		{"Nation", "Supplier", []string{"N_NATIONKEY"}, []string{"S_NATIONKEY"}, false, true},
		{"Nation", "Customer", []string{"N_NATIONKEY"}, []string{"C_NATIONKEY"}, false, true},
		{"Supplier", "PartSupp", []string{"S_SUPPKEY"}, []string{"PS_SUPPKEY"}, false, true},
		{"Part", "PartSupp", []string{"P_PARTKEY"}, []string{"PS_PARTKEY"}, false, true},
		{"PartSupp", "LineItem", []string{"PS_PARTKEY", "PS_SUPPKEY"}, []string{"L_PARTKEY", "L_SUPPKEY"}, false, true},
		// Joins from LineItem directly to Part and Supplier are not part of
		// the official schema but upstream adds them for convenience.
		{"Part", "LineItem", []string{"P_PARTKEY"}, []string{"L_PARTKEY"}, false, true},
		{"Supplier", "LineItem", []string{"S_SUPPKEY"}, []string{"L_SUPPKEY"}, false, true},
	}

	postBuild := func(_ *googlesql.SimpleCatalog, _ *Schema, tables map[string]*googlesql.SimpleTable, _ *googlesql.TypeFactory) error {
		for _, j := range joins {
			t1 := tables[strings.ToLower(j.t1)]
			t2 := tables[strings.ToLower(j.t2)]
			if t1 == nil || t2 == nil {
				return fmt.Errorf("tpch_graph: table not found: %s or %s", j.t1, j.t2)
			}
			if err := addJoinColumn(tf, t1, t2, j.c1, j.c2, j.multi1, j.m2); err != nil {
				return fmt.Errorf("tpch_graph: %s<->%s: %w", j.t1, j.t2, err)
			}
		}
		return nil
	}

	res, err := buildSimple(schema, lo, tf, postBuild)
	if err != nil {
		return nil, err
	}
	// Reflect the new pseudo-columns into the Go-side Schema so DESCRIBE
	// prints them. The catalog SimpleTable already carries them.
	for _, j := range joins {
		if err := mirrorJoinIntoSchema(res.Schema, j.t1, j.t2, j.multi1, j.m2); err != nil {
			return nil, err
		}
	}
	return res, nil
}

// addJoinColumn ports tpch_catalog.cc's AddJoinColumn — it adds one
// pseudo-column on each of `t1` and `t2` whose value is a
// (MULTI)ROW<other-table>.
func addJoinColumn(tf *googlesql.TypeFactory, t1, t2 *googlesql.SimpleTable, names1, names2 []string, isMulti1, isMulti2 bool) error {
	cols1, err := findColumns(t1, names1)
	if err != nil {
		return err
	}
	cols2, err := findColumns(t2, names2)
	if err != nil {
		return err
	}
	t1FullName, err := t1.FullName()
	if err != nil {
		return err
	}
	t2FullName, err := t2.FullName()
	if err != nil {
		return err
	}
	// MakeRowType3 returns a join RowType bound to the source table /
	// columns; its multiRow argument is the cardinality on the *target*
	// side as observed from the column being added.
	rowType2, err := tf.MakeRowType3(t2, t2FullName, isMulti2, cols2, t1, cols1)
	if err != nil {
		return fmt.Errorf("MakeRowType3 for %s: %w", t2FullName, err)
	}
	rowType1, err := tf.MakeRowType3(t1, t1FullName, isMulti1, cols1, t2, cols2)
	if err != nil {
		return fmt.Errorf("MakeRowType3 for %s: %w", t1FullName, err)
	}
	t2Name, err := t2.Name()
	if err != nil {
		return err
	}
	t1Name, err := t1.Name()
	if err != nil {
		return err
	}
	if err := addOnePseudoColumn(t1, t1FullName, makePlural(t2Name, isMulti2), rowType2); err != nil {
		return err
	}
	return addOnePseudoColumn(t2, t2FullName, makePlural(t1Name, isMulti1), rowType1)
}

// addOnePseudoColumn ports tpch_catalog.cc's AddOneJoinColumn,
// including the `Order` -> `Order_` alias workaround for the reserved
// keyword.
func addOnePseudoColumn(tbl *googlesql.SimpleTable, tableFullName, columnName string, typ googlesql.Googlesql_TypeNode) error {
	for {
		col, err := googlesql.NewSimpleColumn(tableFullName, columnName, typ, true /*isPseudoColumn*/, false /*isWritableColumn*/)
		if err != nil {
			return fmt.Errorf("NewSimpleColumn %s.%s: %w", tableFullName, columnName, err)
		}
		if err := tbl.AddColumn(col); err != nil {
			return fmt.Errorf("AddColumn %s.%s: %w", tableFullName, columnName, err)
		}
		if columnName != "Order" {
			return nil
		}
		columnName = "Order_"
	}
}

func findColumns(tbl *googlesql.SimpleTable, names []string) ([]googlesql.Googlesql_ColumnNode, error) {
	out := make([]googlesql.Googlesql_ColumnNode, 0, len(names))
	for _, n := range names {
		c, err := tbl.FindColumnByName(n)
		if err != nil {
			return nil, fmt.Errorf("find column %q: %w", n, err)
		}
		if c == nil {
			return nil, fmt.Errorf("column %q not found", n)
		}
		out = append(out, c)
	}
	return out, nil
}

// makePlural mirrors upstream: append/strip a trailing "s" depending
// on whether the target side is multi-row.
func makePlural(name string, plural bool) string {
	hasS := strings.HasSuffix(name, "s")
	switch {
	case plural && !hasS:
		return name + "s"
	case !plural && hasS:
		return name[:len(name)-1]
	default:
		return name
	}
}

// mirrorJoinIntoSchema appends the same join pseudo-columns to the
// Go-side Schema so DESCRIBE prints them. Kind is left
// TypeKindTypeUnknown — the Schema has no representation for ROW
// types, and DESCRIBE only displays the column name.
func mirrorJoinIntoSchema(schema *Schema, t1Name, t2Name string, isMulti1, isMulti2 bool) error {
	t1, ok := schema.FindTable(t1Name)
	if !ok {
		return fmt.Errorf("schema: table %q not found", t1Name)
	}
	t2, ok := schema.FindTable(t2Name)
	if !ok {
		return fmt.Errorf("schema: table %q not found", t2Name)
	}
	addPseudoCol(t1, makePlural(t2Name, isMulti2))
	addPseudoCol(t2, makePlural(t1Name, isMulti1))
	return nil
}

func addPseudoCol(t *TableSchema, name string) {
	t.Columns = append(t.Columns, ColumnSchema{Name: name, Kind: googlesql.TypeKindTypeUnknown})
	if name == "Order" {
		t.Columns = append(t.Columns, ColumnSchema{Name: "Order_", Kind: googlesql.TypeKindTypeUnknown})
	}
}
