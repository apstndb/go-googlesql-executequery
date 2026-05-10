package webui

import (
	"bytes"
	"fmt"
	"html"
)

// htmlWriter implements executequery.Writer and builds an HTML fragment for the
// right-hand pane (#result). Markup follows google/googlesql page_body.html
// result blocks (div.result, pre.output, code.language-*).
type htmlWriter struct {
	buf    bytes.Buffer
	errMsg string
}

func (h *htmlWriter) setError(msg string) {
	h.errMsg = msg
}

func (h *htmlWriter) Bytes() []byte {
	if h.errMsg != "" {
		return []byte(`<pre id="error" class="error">` + html.EscapeString(h.errMsg) + `</pre>`)
	}
	return h.buf.Bytes()
}

func (h *htmlWriter) StatementText(text string) error {
	h.buf.WriteString(`<div class="result"><h3 class="statement">Statement</h3><pre class="output statement"><code class="language-sql">`)
	h.buf.WriteString(html.EscapeString(text))
	h.buf.WriteString("</code></pre></div>\n")
	return nil
}

func (h *htmlWriter) Parsed(debug string) error {
	return h.writeCodeSection("Parse", "parsed", "language-less", debug)
}

func (h *htmlWriter) Unparsed(sql string) error {
	return h.writeCodeSection("Unparse", "unparsed", "language-sql", sql)
}

func (h *htmlWriter) Resolved(debug string) error {
	return h.writeCodeSection("Analyze", "analyzed", "language-less", debug)
}

func (h *htmlWriter) Described(text string) error {
	return h.writeCodeSection("Describe", "", "language-less", text)
}

func (h *htmlWriter) StartStatement(isFirst bool) error {
	if !isFirst {
		h.buf.WriteString("<hr>\n")
	}
	return nil
}

func (h *htmlWriter) FlushStatement(atEnd bool, errMsg string) error {
	if errMsg != "" {
		h.buf.WriteString(`<pre id="error" class="error">`)
		h.buf.WriteString(html.EscapeString(errMsg))
		h.buf.WriteString("</pre>\n")
	}
	return nil
}

func (h *htmlWriter) writeCodeSection(title, preExtraClass, codeClass, body string) error {
	fmt.Fprintf(&h.buf, `<div class="result"><h3>%s</h3><pre class="output`, html.EscapeString(title))
	if preExtraClass != "" {
		fmt.Fprintf(&h.buf, " %s", html.EscapeString(preExtraClass))
	}
	fmt.Fprintf(&h.buf, `"><code class="%s">`, html.EscapeString(codeClass))
	h.buf.WriteString(html.EscapeString(body))
	h.buf.WriteString("</code></pre></div>\n")
	return nil
}
