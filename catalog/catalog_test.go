package catalog_test

import (
	"strings"
	"testing"

	googlesql "github.com/goccy/go-googlesql"

	"github.com/apstndb/go-googlesql-executequery/cache"
	"github.com/apstndb/go-googlesql-executequery/catalog"
)

func TestMain(m *testing.M) {
	if err := cache.Setup(); err != nil {
		panic(err)
	}
	m.Run()
}

func TestParseName(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		want      catalog.Name
		supported bool
		isErr     bool
	}{
		"":           {catalog.None, true, false},
		"none":       {catalog.None, true, false},
		"sample":     {catalog.Sample, true, false},
		"tpch":       {catalog.TPCH, true, false},
		"tpch_graph": {catalog.TPCHGraph, true, false},
		"bogus":      {"", false, true},
	}
	for in, tc := range cases {
		got, sup, err := catalog.ParseName(in)
		if tc.isErr {
			if err == nil {
				t.Errorf("ParseName(%q): expected error, got %v / %v", in, got, sup)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseName(%q): unexpected error %v", in, err)
			continue
		}
		if got != tc.want || sup != tc.supported {
			t.Errorf("ParseName(%q): got (%q, %v); want (%q, %v)", in, got, sup, tc.want, tc.supported)
		}
	}
}

func TestBuildSampleAndTPCH(t *testing.T) {
	lo, err := googlesql.NewLanguageOptions()
	if err != nil {
		t.Fatalf("NewLanguageOptions: %v", err)
	}
	if err := lo.EnableMaximumLanguageFeatures(); err != nil {
		t.Fatalf("EnableMaximumLanguageFeatures: %v", err)
	}
	tf, err := googlesql.NewTypeFactory()
	if err != nil {
		t.Fatalf("NewTypeFactory: %v", err)
	}

	for _, name := range []catalog.Name{catalog.Sample, catalog.TPCH} {
		t.Run(string(name), func(t *testing.T) {
			res, err := catalog.Build(name, lo, tf)
			if err != nil {
				t.Fatalf("Build(%q): %v", name, err)
			}
			if res.Schema == nil || len(res.Schema.Tables) == 0 {
				t.Errorf("schema empty for %q", name)
			}
			// Verify FindTable is case-insensitive.
			first := res.Schema.Tables[0].Name
			if _, ok := res.Schema.FindTable(strings.ToUpper(first)); !ok {
				t.Errorf("FindTable should be case-insensitive: %q not found", first)
			}
		})
	}
}

func TestTPCHPrimaryKeys(t *testing.T) {
	t.Parallel()
	lo, err := googlesql.NewLanguageOptions()
	if err != nil {
		t.Fatalf("NewLanguageOptions: %v", err)
	}
	tf, err := googlesql.NewTypeFactory()
	if err != nil {
		t.Fatalf("NewTypeFactory: %v", err)
	}
	res, err := catalog.Build(catalog.TPCH, lo, tf)
	if err != nil {
		t.Fatalf("Build(tpch): %v", err)
	}
	want := map[string][]string{
		"Customer": {"C_CUSTKEY"},
		"LineItem": {"L_ORDERKEY", "L_LINENUMBER"},
		"PartSupp": {"PS_PARTKEY", "PS_SUPPKEY"},
	}
	for name, keys := range want {
		ts, ok := res.Schema.FindTable(name)
		if !ok {
			t.Errorf("table %q not in schema", name)
			continue
		}
		if got := strings.Join(ts.PrimaryKey, ","); got != strings.Join(keys, ",") {
			t.Errorf("%s: PrimaryKey = %v; want %v", name, ts.PrimaryKey, keys)
		}
		if !strings.Contains(ts.Format(), "Primary key: ("+strings.Join(keys, ", ")+")") {
			t.Errorf("%s: Format() missing primary key line:\n%s", name, ts.Format())
		}
	}
}

func TestSampleProtoTables(t *testing.T) {
	t.Parallel()
	lo, err := googlesql.NewLanguageOptions()
	if err != nil {
		t.Fatalf("NewLanguageOptions: %v", err)
	}
	if err := lo.EnableMaximumLanguageFeatures(); err != nil {
		t.Fatalf("EnableMaximumLanguageFeatures: %v", err)
	}
	tf, err := googlesql.NewTypeFactory()
	if err != nil {
		t.Fatalf("NewTypeFactory: %v", err)
	}
	res, err := catalog.Build(catalog.Sample, lo, tf)
	if err != nil {
		t.Fatalf("Build(sample): %v", err)
	}

	for _, name := range []string{"TestTable", "EnumTable"} {
		ts, ok := res.Schema.FindTable(name)
		if !ok {
			t.Errorf("schema missing %q", name)
			continue
		}
		if !strings.Contains(ts.Format(), "ENUM<zetasql_test.TestEnum>") {
			t.Errorf("%s.Format() missing enum label:\n%s", name, ts.Format())
		}
	}

	ao, err := googlesql.NewAnalyzerOptions(lo)
	if err != nil {
		t.Fatalf("NewAnalyzerOptions: %v", err)
	}
	// Walk a proto field via dotted notation; this exercises GetProtoField
	// resolution end-to-end through the goccy DescriptorPool wiring.
	out, err := googlesql.AnalyzeStatement(
		"SELECT key, TestEnum, KitchenSink.int64_val, KitchenSink.test_enum FROM TestTable",
		ao, res.Catalog, tf,
	)
	if err != nil {
		t.Fatalf("AnalyzeStatement(TestTable): %v", err)
	}
	resolved, err := out.ResolvedStatement()
	if err != nil {
		t.Fatalf("ResolvedStatement: %v", err)
	}
	dbg, err := resolved.DebugString()
	if err != nil {
		t.Fatalf("DebugString: %v", err)
	}
	if !strings.Contains(dbg, "GetProtoField") {
		t.Errorf("resolved AST missing GetProtoField (proto wiring may be broken):\n%s", dbg)
	}
	if !strings.Contains(dbg, "zetasql_test.KitchenSinkPB") {
		t.Errorf("resolved AST missing proto FQN:\n%s", dbg)
	}
}

func TestBuildTPCHGraph(t *testing.T) {
	lo, err := googlesql.NewLanguageOptions()
	if err != nil {
		t.Fatalf("NewLanguageOptions: %v", err)
	}
	if err := lo.EnableMaximumLanguageFeatures(); err != nil {
		t.Fatalf("EnableMaximumLanguageFeatures: %v", err)
	}
	tf, err := googlesql.NewTypeFactory()
	if err != nil {
		t.Fatalf("NewTypeFactory: %v", err)
	}
	res, err := catalog.Build(catalog.TPCHGraph, lo, tf)
	if err != nil {
		t.Fatalf("Build(tpch_graph): %v", err)
	}
	cust, ok := res.Schema.FindTable("Customer")
	if !ok {
		t.Fatalf("Customer not found in schema")
	}
	// The graph variant adds Orders (MULTIROW<Orders>) and Nation
	// (ROW<Nation>) pseudo-columns to Customer. Match by name only;
	// Kind is unknown for ROW types in our Schema (we expose
	// TypeKindUnknown).
	hasOrders := false
	hasNation := false
	for _, c := range cust.Columns {
		switch c.Name {
		case "Orders":
			hasOrders = true
		case "Nation":
			hasNation = true
		}
	}
	if !hasOrders || !hasNation {
		t.Errorf("Customer columns missing Orders/Nation pseudo-columns: %+v", cust.Columns)
	}
}
