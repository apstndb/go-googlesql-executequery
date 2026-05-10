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

func TestPageTemplateUpstreamOnlyModesUseHiddenUncheckedControls(t *testing.T) {
	t.Parallel()
	if !strings.Contains(pageTemplate, `webui-upcoming-tool-modes`) {
		t.Fatal("expected wrapper for upcoming tool modes")
	}
	for _, pair := range []struct{ val, kind string }{
		{"execute", "checkbox"},
		{"explain", "checkbox"},
		{"unanalyze", "checkbox"},
	} {
		// Require checkbox inputs with name=mode (unchecked); not type=hidden.
		if !strings.Contains(pageTemplate, `<input type="`+pair.kind+`" name="mode" value="`+pair.val+`"`) {
			t.Fatalf("expected checkbox name=mode value=%s", pair.val)
		}
	}
	if strings.Contains(pageTemplate, `type="hidden" name="mode"`) {
		t.Fatal("do not use hidden inputs named mode — use unchecked checkboxes inside a hidden wrapper")
	}
	if !strings.Contains(pageTemplate, `webui-upcoming-target-syntax-pipe`) {
		t.Fatal("expected wrapper label for pipe target syntax")
	}
	if !strings.Contains(pageTemplate, `<input type="radio" name="target_syntax_mode" value="pipe"`) {
		t.Fatal("expected pipe radio in group with standard (standard stays checked; pipe not submitted)")
	}
}
