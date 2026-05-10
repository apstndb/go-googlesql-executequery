package webui

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
    <textarea id="sql" name="sql" placeholder="SELECT 1 AS col1, 'hello' AS col2;" required></textarea>
  </div>
  <div class="form-group">
    <label>Mode</label>
    <div class="checkbox-group">
      <label><input type="checkbox" name="mode" value="parse" checked> parse</label>
      <label><input type="checkbox" name="mode" value="unparse"> unparse</label>
      <label><input type="checkbox" name="mode" value="analyze" checked> analyze</label>
    </div>
  </div>
  <div class="form-group">
    <label for="catalog">Catalog</label>
    <select id="catalog" name="catalog">
      <option value="none">none</option>
      <option value="sample" selected>sample</option>
      <option value="tpch">tpch</option>
      <option value="tpch_graph">tpch_graph</option>
    </select>
  </div>
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
    const response = await fetch('/run', {
      method: 'POST',
      body: new FormData(form)
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
