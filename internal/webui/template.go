package webui

import (
	_ "embed"
	"html/template"
)

// page_style.css is a verbatim copy of google/googlesql tools/execute_query/web/style.css
// (Apache-2.0); kept in-repo for go:embed so the Go UI matches upstream layout/CSS without a civetweb dependency.
//
//go:embed page_style.css
var pageStyleCSS string

// indexData mirrors google/googlesql tools/execute_query/web/page_body.html field names and structure.
type indexData struct {
	ToolModes         []toolModeRow
	Catalogs          []selectOpt
	SQLModes          []radioRow
	TargetSyntaxModes []radioRow // standard only; Pipe is a separate span in the template (unsupported)
	LanguageFeatures  []selectOpt
	ASTRewrites       []selectOpt
}

type toolModeRow struct {
	Value   string
	Label   string
	ID      string
	Checked bool
	Hidden  bool
}

type radioRow struct {
	Value   string
	Label   string
	ID      string
	Checked bool
}

type selectOpt struct {
	Value    string
	Label    string
	Selected bool
}

// pageData adds embedded stylesheet bytes (same roles as upstream page_template.html {{{css}}}).
type pageData struct {
	Style template.CSS
	indexData
}

func defaultIndexData() indexData {
	return indexData{
		ToolModes: []toolModeRow{
			{Value: "execute", Label: "Execute", ID: "mode-execute", Checked: false, Hidden: true},
			{Value: "analyze", Label: "Analyze", ID: "mode-analyze", Checked: true, Hidden: false},
			{Value: "parse", Label: "Parse", ID: "mode-parse", Checked: true, Hidden: false},
			{Value: "explain", Label: "Explain", ID: "mode-explain", Checked: false, Hidden: true},
			{Value: "unanalyze", Label: "Unanalyze", ID: "mode-unanalyze", Checked: false, Hidden: true},
			{Value: "unparse", Label: "Unparse", ID: "mode-unparse", Checked: false, Hidden: false},
		},
		Catalogs: []selectOpt{
			{Value: "none", Label: "none"},
			{Value: "sample", Label: "sample", Selected: true},
			{Value: "tpch", Label: "tpch"},
			{Value: "tpch_graph", Label: "tpch_graph"},
		},
		SQLModes: []radioRow{
			{Value: "query", Label: "Query", ID: "sql-mode-query", Checked: true},
			{Value: "expression", Label: "Expression", ID: "sql-mode-expression", Checked: false},
			{Value: "script", Label: "Script", ID: "sql-mode-script", Checked: false},
		},
		TargetSyntaxModes: []radioRow{
			{Value: "standard", Label: "Standard", ID: "target-syntax-mode-standard", Checked: true},
		},
		LanguageFeatures: []selectOpt{
			{Value: "NONE", Label: "NONE"},
			{Value: "MAXIMUM", Label: "MAXIMUM", Selected: true},
			{Value: "ALL", Label: "ALL"},
			{Value: "ALL_MINUS_DEV", Label: "ALL_MINUS_DEV"},
			{Value: "DEFAULTS", Label: "DEFAULTS"},
			{Value: "DEFAULTS_MINUS_DEV", Label: "DEFAULTS_MINUS_DEV"},
		},
		ASTRewrites: []selectOpt{
			{Value: "", Label: "(default)", Selected: true},
			{Value: "NONE", Label: "NONE"},
			{Value: "ALL", Label: "ALL"},
			{Value: "ALL_MINUS_DEV", Label: "ALL_MINUS_DEV"},
			{Value: "DEFAULTS", Label: "DEFAULTS"},
			{Value: "DEFAULTS_MINUS_DEV", Label: "DEFAULTS_MINUS_DEV"},
		},
	}
}

const pageTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>GoogleSQL Execute Query</title>
<style>{{.Style}}</style>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/stackoverflow-light.min.css">
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/sql.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/less.min.js"></script>
<script>hljs.highlightAll();</script>
</head>
<body>
<main>
  <div id="header">
    <a href="/">GoogleSQL Execute Query</a>
  </div>
  <div class="left-section">
    <div id="editor"></div>
    <form id="form" action="/run" method="post">
      <textarea id="query" tabindex="0" name="query" spellcheck="false" placeholder="Enter your query here..."></textarea>
      <div class="options">
        <div>
          <fieldset>
            <legend>Mode</legend>
            {{- range .ToolModes}}
            <span class="mode-option">
              <input type="checkbox" name="mode" value="{{.Value}}" id="{{.ID}}"{{if .Checked}} checked{{end}}{{if .Hidden}} hidden{{end}}>
              <label for="{{.ID}}">{{.Label}}</label>
            </span>
            {{- end}}
          </fieldset>
          <span class="select-options">
            <label for="catalog-select">Catalog:</label>
            <select name="catalog" id="catalog-select">
              {{- range .Catalogs}}
              <option value="{{.Value}}"{{if .Selected}} selected{{end}}>{{.Label}}</option>
              {{- end}}
            </select>
          </span>
          <details>
            <summary>Advanced Options</summary>
            <fieldset>
              <legend>SQL Mode</legend>
              {{- range .SQLModes}}
              <input type="radio" name="sql_mode" value="{{.Value}}" id="{{.ID}}"{{if .Checked}} checked{{end}}>
              <label for="{{.ID}}">{{.Label}}</label>
              {{- end}}
            </fieldset>
            <fieldset>
              <legend>Unanalyze Syntax Mode</legend>
              {{- range .TargetSyntaxModes}}
              <input type="radio" name="target_syntax_mode" value="{{.Value}}" id="{{.ID}}"{{if .Checked}} checked{{end}}>
              <label for="{{.ID}}">{{.Label}}</label>
              {{- end}}
              <span hidden class="webui-upcoming-target-syntax-pipe">
                <input type="radio" name="target_syntax_mode" value="pipe" id="target-syntax-mode-pipe">
                <label for="target-syntax-mode-pipe">Pipe</label>
              </span>
            </fieldset>
            <span class="select-options">
              <label for="language-features-select">Enabled Language Features:</label>
              <select name="language-features" id="language-features-select">
                {{- range .LanguageFeatures}}
                <option value="{{.Value}}"{{if .Selected}} selected{{end}}>{{.Label}}</option>
                {{- end}}
              </select>
            </span>
            <span class="select-options">
              <label for="ast-rewrites-select">Enabled AST Rewrites:</label>
              <select name="ast-rewrites" id="ast-rewrites-select">
                {{- range .ASTRewrites}}
                <option value="{{.Value}}"{{if .Selected}} selected{{end}}>{{.Label}}</option>
                {{- end}}
              </select>
            </span>
          </details>
        </div>
        <input type="submit" id="submit" value="Submit">
      </div>
    </form>
  </div>
  <div class="right-section" id="statements">
    <div id="result"></div>
  </div>
</main>
<script>
document.getElementById('form').addEventListener('submit', async function(e) {
  e.preventDefault();
  const form = e.target;
  const result = document.getElementById('result');
  result.innerHTML = '<p>Running...</p>';
  try {
    const fd = new FormData(form);
    const params = new URLSearchParams();
    for (const pair of fd.entries()) {
      params.append(pair[0], pair[1]);
    }
    const response = await fetch('/run', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
      body: params.toString(),
    });
    const html = await response.text();
    result.innerHTML = html;
    if (window.hljs) {
      result.querySelectorAll('pre code').forEach(function(el) { hljs.highlightElement(el); });
    }
  } catch (err) {
    result.innerHTML = '<pre id="error" class="error">' + err.message + '</pre>';
  }
});
</script>
</body>
</html>
`

var tmpl = template.Must(template.New("webui").Parse(pageTemplate))
