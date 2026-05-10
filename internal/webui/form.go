package webui

import (
	"net/http"
	"strings"

	executequery "github.com/apstndb/go-googlesql-executequery"
)

// configFromForm builds executequery.Config from POST parameters (names match
// the embedded page template in template.go: textarea name="query", etc.).
func configFromForm(r *http.Request) (executequery.Config, error) {
	var cfg executequery.Config

	cfg.CatalogName = strings.TrimSpace(r.FormValue("catalog"))

	for _, m := range r.Form["mode"] {
		mode, ok := executequery.ParseMode(m)
		if ok {
			cfg.Modes = append(cfg.Modes, mode)
		}
	}
	if len(cfg.Modes) == 0 {
		cfg.Modes = []executequery.Mode{executequery.ModeAnalyze}
	}

	if v := strings.TrimSpace(r.FormValue("sql_mode")); v != "" {
		if sm, ok := executequery.ParseSQLMode(v); ok {
			cfg.SQLMode = sm
		}
	}

	if v := strings.TrimSpace(r.FormValue("target_syntax_mode")); v != "" {
		switch strings.ToLower(v) {
		case "standard":
			cfg.TargetSyntax = ""
		case "pipe":
			cfg.TargetSyntax = "pipe"
		default:
			cfg.TargetSyntax = v
		}
	}

	if v := strings.TrimSpace(r.FormValue("language-features")); v != "" {
		fs, err := executequery.ParseFeatureSet(v)
		if err != nil {
			return executequery.Config{}, err
		}
		cfg.EnabledLanguageFeatures = fs
	}

	if v := strings.TrimSpace(r.FormValue("ast-rewrites")); v != "" {
		rs, err := executequery.ParseRewriteSet(v)
		if err != nil {
			return executequery.Config{}, err
		}
		cfg.EnabledASTRewrites = rs
	}

	return cfg, nil
}
