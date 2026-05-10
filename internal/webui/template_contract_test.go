package webui

import (
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

func TestPageTemplateHiddenUpstreamParityRefs(t *testing.T) {
	t.Parallel()
	for _, needle := range []string{
		`type="hidden"`,
		`value="execute"`,
		`value="explain"`,
		`value="unanalyze"`,
		`value="pipe"`,
	} {
		if !strings.Contains(pageTemplate, needle) {
			t.Fatalf("expected hidden parity marker %q in pageTemplate", needle)
		}
	}
	// Unsupported parity refs must omit name so URLSearchParams POST is unchanged.
	if strings.Contains(pageTemplate, `type="hidden" name="mode"`) {
		t.Fatal("hidden upstream refs must not use name=mode (would pollute POST)")
	}
	if strings.Contains(pageTemplate, `type="hidden" name="target_syntax_mode"`) {
		t.Fatal("hidden pipe ref must not use name=target_syntax_mode")
	}
}
