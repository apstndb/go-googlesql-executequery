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
	for _, val := range []string{"execute", "explain", "unanalyze"} {
		needle := `<input type="checkbox" name="mode" value="` + val + `" hidden`
		if !strings.Contains(pageTemplate, needle) {
			t.Fatalf("expected hidden checkbox name=mode value=%s", val)
		}
	}
	if strings.Contains(pageTemplate, `type="hidden" name="mode"`) {
		t.Fatal("do not use type=hidden for mode — use unchecked checkbox with hidden attribute")
	}
	if !strings.Contains(pageTemplate, `<input type="radio" name="target_syntax_mode" value="pipe" hidden`) {
		t.Fatal("expected hidden pipe radio (standard selected; pipe not submitted)")
	}
	if strings.Contains(pageTemplate, `webui-upcoming-tool-modes`) {
		t.Fatal("drop redundant wrapper div — hidden lives on each control")
	}
}
