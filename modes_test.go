package executequery_test

import (
	"testing"

	executequery "github.com/apstndb/go-googlesql-executequery"
)

func TestParseMode(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		want executequery.Mode
		ok   bool
	}{
		"parse":       {executequery.ModeParse, true},
		"unparse":     {executequery.ModeUnparse, true},
		"analyze":     {executequery.ModeAnalyze, true},
		"resolve":     {executequery.ModeAnalyze, true},
		"unanalyze":   {executequery.ModeUnanalyze, true},
		"sql_builder": {executequery.ModeUnanalyze, true},
		"explain":     {executequery.ModeExplain, true},
		"execute":     {executequery.ModeExecute, true},
		"bogus":       {"", false},
	}
	for in, want := range cases {
		got, ok := executequery.ParseMode(in)
		if ok != want.ok {
			t.Errorf("ParseMode(%q) ok: got %v, want %v", in, ok, want.ok)
		}
		if got != want.want {
			t.Errorf("ParseMode(%q) got %v, want %v", in, got, want.want)
		}
	}
}

func TestModeIsSupported(t *testing.T) {
	t.Parallel()
	cases := map[executequery.Mode]bool{
		executequery.ModeParse:     true,
		executequery.ModeUnparse:   true,
		executequery.ModeAnalyze:   true,
		executequery.ModeUnanalyze: false,
		executequery.ModeExplain:   false,
		executequery.ModeExecute:   false,
	}
	for m, want := range cases {
		if got := m.IsSupported(); got != want {
			t.Errorf("(%s).IsSupported() = %v, want %v", m, got, want)
		}
	}
}
