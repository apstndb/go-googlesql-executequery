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
		"tpch_graph": {catalog.TPCHGraph, false, false},
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

func TestBuildTPCHGraphRejected(t *testing.T) {
	lo, _ := googlesql.NewLanguageOptions()
	tf, _ := googlesql.NewTypeFactory()
	if _, err := catalog.Build(catalog.TPCHGraph, lo, tf); err == nil {
		t.Errorf("expected error for tpch_graph, got nil")
	}
}
