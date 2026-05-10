package webui

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	executequery "github.com/apstndb/go-googlesql-executequery"
)

func TestNormalizeLanguageFeaturesChoice(t *testing.T) {
	t.Parallel()
	if got := normalizeLanguageFeaturesChoice("MAXIMUM"); got != "ALL_MINUS_DEV" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeLanguageFeaturesChoice("NONE"); got != "NONE" {
		t.Fatalf("got %q", got)
	}
}

func TestConfigFromFormAdvanced(t *testing.T) {
	t.Parallel()
	form := url.Values{}
	form.Set("catalog", "tpch")
	form.Add("mode", "parse")
	form.Set("sql_mode", "expression")
	form.Set("target_syntax_mode", "pipe")
	form.Set("language-features", "NONE")
	form.Set("ast-rewrites", "NONE")

	req, err := http.NewRequest(http.MethodPost, "/run", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	cfg, err := configFromForm(req)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CatalogName != "tpch" {
		t.Fatalf("catalog: %q", cfg.CatalogName)
	}
	if len(cfg.Modes) != 1 || cfg.Modes[0] != executequery.ModeParse {
		t.Fatalf("modes: %+v", cfg.Modes)
	}
	if cfg.SQLMode != executequery.SQLModeExpression {
		t.Fatalf("sql_mode: %v", cfg.SQLMode)
	}
	if cfg.TargetSyntax != "pipe" {
		t.Fatalf("target_syntax: %q", cfg.TargetSyntax)
	}
	if err := cfg.Validate(); err == nil || !errors.Is(err, executequery.ErrUnsupportedFlag) {
		t.Fatalf("expected unsupported target_syntax pipe: %v", err)
	}
}
