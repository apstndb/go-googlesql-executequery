package webui

import (
	"bytes"
	"fmt"
	"html"
)

// htmlWriter implements executequery.Writer and builds an HTML
// fragment suitable for inserting into the web UI result pane.
type htmlWriter struct {
	buf    bytes.Buffer
	errMsg string
}

func (h *htmlWriter) setError(msg string) {
	h.errMsg = msg
}

func (h *htmlWriter) Bytes() []byte {
	if h.errMsg != "" {
		return []byte(`<div class="result-error">` + html.EscapeString(h.errMsg) + `</div>`)
	}
	return h.buf.Bytes()
}

func (h *htmlWriter) StatementText(text string) error {
	h.buf.WriteString(`<div class="result-section"><h3>Statement</h3><pre>`)
	h.buf.WriteString(html.EscapeString(text))
	h.buf.WriteString("</pre></div>\n")
	return nil
}

func (h *htmlWriter) Parsed(debug string) error {
	return h.writeSection("Parse", debug)
}

func (h *htmlWriter) Unparsed(sql string) error {
	return h.writeSection("Unparse", sql)
}

func (h *htmlWriter) Resolved(debug string) error {
	return h.writeSection("Analyze", debug)
}

func (h *htmlWriter) Described(text string) error {
	return h.writeSection("Describe", text)
}

func (h *htmlWriter) StartStatement(isFirst bool) error {
	if !isFirst {
		h.buf.WriteString("<hr>\n")
	}
	return nil
}

func (h *htmlWriter) FlushStatement(atEnd bool, errMsg string) error {
	if errMsg != "" {
		h.buf.WriteString(`<div class="result-error">`)
		h.buf.WriteString(html.EscapeString(errMsg))
		h.buf.WriteString("</div>\n")
	}
	return nil
}

func (h *htmlWriter) writeSection(title, body string) error {
	fmt.Fprintf(&h.buf, `<div class="result-section"><h3>%s</h3><pre>`, html.EscapeString(title))
	h.buf.WriteString(html.EscapeString(body))
	h.buf.WriteString("</pre></div>\n")
	return nil
}
