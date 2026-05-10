package webui

import (
	"html/template"
	"strings"
	"testing"
)

// The upstream C++ web UI posts with a plain HTML form (application/x-www-form-urlencoded).
// See third_party/googlesql/googlesql/tools/execute_query/web/page_body.html (method="post").
// This bundle keeps one interactive fetch() path but must serialize like a native form POST.
func TestPageTemplateUsesUrlEncodedFetchBody(t *testing.T) {
	t.Parallel()
	if !strings.Contains(pageTemplate, "application/x-www-form-urlencoded") {
		t.Fatal("expected fetch() to set Content-Type to application/x-www-form-urlencoded")
	}
	if !strings.Contains(pageTemplate, "new URLSearchParams()") {
		t.Fatal("expected URLSearchParams built from FormData entries")
	}
	if strings.Contains(pageTemplate, "body: new FormData(form)") {
		t.Fatal("do not pass FormData directly to fetch(); it forces multipart/form-data")
	}
}

func TestPageTemplateMatchesUpstreamPageShape(t *testing.T) {
	t.Parallel()
	for _, needle := range []string{
		`<main>`,
		`id="header"`,
		`class="left-section"`,
		`id="form"`,
		`id="query"`,
		`name="query"`,
		`class="right-section"`,
		`id="statements"`,
		`id="catalog-select"`,
		`id="language-features-select"`,
		`hljs.highlightAll`,
	} {
		if !strings.Contains(pageTemplate, needle) {
			t.Fatalf("expected upstream-shaped marker %q in pageTemplate", needle)
		}
	}
}

func TestPageTemplateUpstreamOnlyModesUseHiddenUncheckedControls(t *testing.T) {
	t.Parallel()
	var buf strings.Builder
	if err := tmpl.Execute(&buf, pageData{
		Style:     template.CSS(""),
		indexData: defaultIndexData(),
	}); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, val := range []string{"execute", "explain", "unanalyze"} {
		needle := `<input type="checkbox" name="mode" value="` + val + `" id="mode-` + val + `" hidden`
		if !strings.Contains(html, needle) {
			t.Fatalf("expected hidden checkbox name=mode value=%s", val)
		}
	}
	if strings.Contains(html, `type="hidden" name="mode"`) {
		t.Fatal("do not use type=hidden for mode — use unchecked checkbox with hidden attribute")
	}
	if !strings.Contains(html, `webui-upcoming-target-syntax-pipe`) {
		t.Fatal("expected unsupported pipe option wrapper")
	}
}
