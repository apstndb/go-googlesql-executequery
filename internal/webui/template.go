package webui

import "html/template"

// indexData drives page_body-style markup (google/googlesql page_body.html).
type indexData struct {
	ToolModes         []toolModeRow
	Catalogs          []selectOpt
	SQLModes          []radioRow
	TargetSyntaxModes []radioRow
	LanguageFeatures  []selectOpt
	ASTRewrites       []selectOpt
}

type toolModeRow struct {
	Value   string
	Label   string
	Checked bool
}

type radioRow struct {
	Value   string
	Label   string
	Checked bool
}

type selectOpt struct {
	Value    string
	Label    string
	Selected bool
}

func defaultIndexData() indexData {
	return indexData{
		ToolModes: []toolModeRow{
			{Value: "execute", Label: "execute", Checked: false},
			{Value: "analyze", Label: "analyze", Checked: true},
			{Value: "parse", Label: "parse", Checked: true},
			{Value: "explain", Label: "explain", Checked: false},
			{Value: "unanalyze", Label: "unanalyze", Checked: false},
			{Value: "unparse", Label: "unparse", Checked: false},
		}, // order matches google/googlesql page_body.html
		Catalogs: []selectOpt{
			{Value: "none", Label: "none"},
			{Value: "sample", Label: "sample", Selected: true},
			{Value: "tpch", Label: "tpch"},
			{Value: "tpch_graph", Label: "tpch_graph"},
		},
		SQLModes: []radioRow{
			{Value: "query", Label: "Query", Checked: true},
			{Value: "expression", Label: "Expression", Checked: false},
			{Value: "script", Label: "Script", Checked: false},
		},
		TargetSyntaxModes: []radioRow{
			{Value: "standard", Label: "Standard", Checked: true},
			{Value: "pipe", Label: "Pipe", Checked: false},
		},
		LanguageFeatures: []selectOpt{
			{Value: "NONE", Label: "NONE"},
			{Value: "MAXIMUM", Label: "MAXIMUM (maps to ALL_MINUS_DEV)", Selected: true},
			{Value: "ALL", Label: "ALL"},
			{Value: "ALL_MINUS_DEV", Label: "ALL_MINUS_DEV"},
			{Value: "DEFAULTS", Label: "DEFAULTS"},
			{Value: "DEFAULTS_MINUS_DEV", Label: "DEFAULTS_MINUS_DEV"},
		},
		ASTRewrites: []selectOpt{
			{Value: "", Label: "(default analyzer rewrites)", Selected: true},
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
<style>
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  background: #f5f5f5;
}
.container {
  background: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
h1 {
  margin-top: 0;
  color: #333;
}
.form-group {
  margin-bottom: 16px;
}
label {
  display: block;
  font-weight: 600;
  margin-bottom: 4px;
  color: #555;
}
textarea {
  width: 100%;
  min-height: 120px;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: "SF Mono", Monaco, "Cascadia Code", monospace;
  font-size: 14px;
  resize: vertical;
}
select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  font-size: 14px;
}
.checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}
.checkbox-group label {
  display: inline;
  font-weight: normal;
  cursor: pointer;
}
.radio-group label {
  display: inline;
  font-weight: normal;
  margin-right: 16px;
  cursor: pointer;
}
fieldset {
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 12px;
}
legend {
  font-weight: 600;
  color: #555;
}
details {
  margin-top: 8px;
}
button {
  background: #1a73e8;
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
}
button:hover {
  background: #1557b0;
}
#result {
  margin-top: 24px;
  padding: 16px;
  background: #fafafa;
  border-radius: 4px;
  border: 1px solid #e0e0e0;
}
.result-section {
  margin-bottom: 16px;
}
.result-section h3 {
  margin: 0 0 8px 0;
  font-size: 14px;
  text-transform: uppercase;
  color: #666;
  letter-spacing: 0.5px;
}
.result-section pre {
  margin: 0;
  padding: 12px;
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.5;
}
.result-error {
  padding: 12px;
  background: #fce8e6;
  border: 1px solid #f5c6cb;
  border-radius: 4px;
  color: #721c24;
}
hr {
  border: none;
  border-top: 1px solid #e0e0e0;
  margin: 16px 0;
}
</style>
</head>
<body>
<div class="container">
<h1>GoogleSQL Execute Query</h1>
<form id="queryForm" action="/run" method="post">
  <div class="form-group">
    <label for="sql">SQL</label>
    <textarea id="sql" name="sql" placeholder="SELECT 1 AS col1, 'hello' AS col2;"></textarea>
  </div>
  <div class="form-group">
    <fieldset>
      <legend>Mode</legend>
      <div class="checkbox-group">
        {{- range .ToolModes}}
        <label><input type="checkbox" name="mode" value="{{.Value}}"{{if .Checked}} checked{{end}}> {{.Label}}</label>
        {{- end}}
      </div>
    </fieldset>
  </div>
  <div class="form-group">
    <label for="catalog">Catalog</label>
    <select id="catalog" name="catalog">
      {{- range .Catalogs}}
      <option value="{{.Value}}"{{if .Selected}} selected{{end}}>{{.Label}}</option>
      {{- end}}
    </select>
  </div>
  <details>
    <summary>Advanced Options</summary>
    <div class="form-group">
      <fieldset>
        <legend>SQL Mode</legend>
        <div class="radio-group">
          {{- range .SQLModes}}
          <label><input type="radio" name="sql_mode" value="{{.Value}}"{{if .Checked}} checked{{end}}> {{.Label}}</label>
          {{- end}}
        </div>
      </fieldset>
    </div>
    <div class="form-group">
      <fieldset>
        <legend>Unanalyze Syntax Mode</legend>
        <div class="radio-group">
          {{- range .TargetSyntaxModes}}
          <label><input type="radio" name="target_syntax_mode" value="{{.Value}}"{{if .Checked}} checked{{end}}> {{.Label}}</label>
          {{- end}}
        </div>
      </fieldset>
    </div>
    <div class="form-group">
      <label for="language-features">Enabled Language Features</label>
      <select name="language-features" id="language-features">
        {{- range .LanguageFeatures}}
        <option value="{{.Value}}"{{if .Selected}} selected{{end}}>{{.Label}}</option>
        {{- end}}
      </select>
    </div>
    <div class="form-group">
      <label for="ast-rewrites">Enabled AST Rewrites</label>
      <select name="ast-rewrites" id="ast-rewrites">
        {{- range .ASTRewrites}}
        <option value="{{.Value}}"{{if .Selected}} selected{{end}}>{{.Label}}</option>
        {{- end}}
      </select>
    </div>
  </details>
  <button type="submit">Run</button>
</form>
<div id="result"></div>
</div>
<script>
document.getElementById('queryForm').addEventListener('submit', async function(e) {
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
  } catch (err) {
    result.innerHTML = '<div class="result-error">' + err.message + '</div>';
  }
});
</script>
</body>
</html>
`

var tmpl = template.Must(template.New("webui").Parse(pageTemplate))
