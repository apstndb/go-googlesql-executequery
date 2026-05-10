package executequery_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	executequery "github.com/apstndb/go-googlesql-executequery"
	"github.com/apstndb/go-googlesql-executequery/cache"
)

func TestMain(m *testing.M) {
	if err := cache.Setup(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type captureWriter struct {
	parsed   []string
	unparsed []string
	resolved []string
	desc     []string
	stmts    int
}

func (c *captureWriter) StatementText(_ string) error { return nil }
func (c *captureWriter) Parsed(s string) error        { c.parsed = append(c.parsed, s); return nil }
func (c *captureWriter) Unparsed(s string) error      { c.unparsed = append(c.unparsed, s); return nil }
func (c *captureWriter) Resolved(s string) error      { c.resolved = append(c.resolved, s); return nil }
func (c *captureWriter) Described(s string) error     { c.desc = append(c.desc, s); return nil }
func (c *captureWriter) StartStatement(_ bool) error  { c.stmts++; return nil }
func (c *captureWriter) FlushStatement(_ bool, _ string) error {
	return nil
}

func TestRunSupportedModes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		cfg      executequery.Config
		sql      string
		wantSubs []string // expected to appear in any of the captured outputs
	}{
		{
			name:     "parse",
			cfg:      executequery.Config{Modes: []executequery.Mode{executequery.ModeParse}},
			sql:      "SELECT 1+1",
			wantSubs: []string{"QueryStatement [", "BinaryExpression"},
		},
		{
			name:     "unparse",
			cfg:      executequery.Config{Modes: []executequery.Mode{executequery.ModeUnparse}},
			sql:      "SELECT  1 +  1",
			wantSubs: []string{"SELECT"},
		},
		{
			name:     "analyze",
			cfg:      executequery.Config{Modes: []executequery.Mode{executequery.ModeAnalyze}},
			sql:      "SELECT 1 AS x",
			wantSubs: []string{"QueryStmt", "Literal", "INT64"},
		},
		{
			name:     "expression mode",
			cfg:      executequery.Config{Modes: []executequery.Mode{executequery.ModeAnalyze}, SQLMode: executequery.SQLModeExpression},
			sql:      "1 + 2",
			wantSubs: []string{"INT64"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := &captureWriter{}
			if err := executequery.Run(context.Background(), tc.sql, tc.cfg, w); err != nil {
				t.Fatalf("Run: %v", err)
			}
			joined := strings.Join(append(append(append(w.parsed, w.unparsed...), w.resolved...), w.desc...), "\n")
			for _, sub := range tc.wantSubs {
				if !strings.Contains(joined, sub) {
					t.Errorf("expected %q in output, got:\n%s", sub, joined)
				}
			}
		})
	}
}

func TestRunUnsupportedMode(t *testing.T) {
	t.Parallel()
	cfg := executequery.Config{Modes: []executequery.Mode{executequery.ModeExecute}}
	w := &captureWriter{}
	err := executequery.Run(context.Background(), "SELECT 1", cfg, w)
	if !errors.Is(err, executequery.ErrUnsupportedMode) {
		t.Fatalf("expected ErrUnsupportedMode, got %v", err)
	}
	if !strings.Contains(err.Error(), executequery.ReasonModeExecute) {
		t.Errorf("error message should include reason, got %q", err)
	}
}

func TestRunUnsupportedFlag(t *testing.T) {
	t.Parallel()
	cfg := executequery.Config{
		Modes:      []executequery.Mode{executequery.ModeAnalyze},
		OutputMode: "json",
	}
	w := &captureWriter{}
	err := executequery.Run(context.Background(), "SELECT 1", cfg, w)
	if !errors.Is(err, executequery.ErrUnsupportedFlag) {
		t.Fatalf("expected ErrUnsupportedFlag, got %v", err)
	}
}

func TestRunCatalogTPCH(t *testing.T) {
	t.Parallel()
	cfg := executequery.Config{
		Modes:       []executequery.Mode{executequery.ModeAnalyze},
		CatalogName: "tpch",
	}
	w := &captureWriter{}
	if err := executequery.Run(context.Background(), "SELECT count(*) FROM Orders", cfg, w); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(w.resolved) == 0 {
		t.Fatalf("expected resolved output")
	}
	if !strings.Contains(w.resolved[0], "Orders") {
		t.Errorf("resolved output should reference Orders: %s", w.resolved[0])
	}
}

func TestRunDescribe(t *testing.T) {
	t.Parallel()
	cfg := executequery.Config{
		Modes:       []executequery.Mode{executequery.ModeAnalyze},
		CatalogName: "tpch",
	}
	w := &captureWriter{}
	if err := executequery.Run(context.Background(), "DESCRIBE Orders", cfg, w); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(w.desc) != 1 {
		t.Fatalf("expected one DESCRIBE emit, got %d", len(w.desc))
	}
	if !strings.Contains(w.desc[0], "O_ORDERKEY") {
		t.Errorf("describe output missing column: %s", w.desc[0])
	}
}

func TestRunMultiStatement(t *testing.T) {
	t.Parallel()
	cfg := executequery.Config{Modes: []executequery.Mode{executequery.ModeParse}}
	w := &captureWriter{}
	if err := executequery.Run(context.Background(), "SELECT 1; SELECT 2;", cfg, w); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(w.parsed) != 2 {
		t.Fatalf("expected 2 parsed emits, got %d (output: %#v)", len(w.parsed), w.parsed)
	}
}

func TestRunCatalogTPCHGraph(t *testing.T) {
	t.Parallel()
	fs, err := executequery.ParseFeatureSet("ALL_MINUS_DEV,+FEATURE_ROW_TYPE")
	if err != nil {
		t.Fatalf("ParseFeatureSet: %v", err)
	}
	cfg := executequery.Config{
		Modes:                   []executequery.Mode{executequery.ModeAnalyze},
		CatalogName:             "tpch_graph",
		EnabledLanguageFeatures: fs,
	}
	w := &captureWriter{}
	if err := executequery.Run(context.Background(), "SELECT c.C_NAME FROM Customer c, c.Orders o LIMIT 1", cfg, w); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(w.resolved) != 1 {
		t.Fatalf("expected 1 resolved emit, got %d (%v)", len(w.resolved), w.resolved)
	}
	if !strings.Contains(w.resolved[0], "Customer.C_NAME") {
		t.Errorf("resolved AST missing Customer.C_NAME: %s", w.resolved[0])
	}
}
