//go:build integration

package webui_test

// Integration tests exercise the same POST encoding as the embedded page script:
// google/googlesql serves a classic HTML form POST (application/x-www-form-urlencoded)
// from page_body.html; this CLI uses fetch() but serializes FormData into URLSearchParams
// so servers see the same encoding as a native form submit.

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"github.com/apstndb/go-googlesql-executequery/internal/webui"
)

func TestWebUIHeadlessFormSubmit(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped with -short")
	}

	srv := webui.NewServer(0)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	page := ts.URL + "/"
	var resultHTML string

	// Mirror internal/webui/template.go submit handler: FormData → URLSearchParams,
	// Content-Type application/x-www-form-urlencoded (not multipart FormData).
	submitJS := `
(async () => {
  const form = document.getElementById('form');
  const result = document.getElementById('result');
  result.innerHTML = '<p>Running...</p>';
  document.querySelector('#query').value = 'SELECT 1 AS x';
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
  result.innerHTML = await response.text();
})()
`

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(page),
		chromedp.WaitVisible(`#query`, chromedp.ByQuery),
		chromedp.Evaluate(submitJS, nil, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
		chromedp.InnerHTML(`#result`, &resultHTML, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("chromedp: %v", err)
	}

	if !strings.Contains(resultHTML, `class="result"`) {
		t.Fatalf("missing result sections in %#q", truncate(resultHTML, 800))
	}
	if !strings.Contains(resultHTML, "Parse") {
		t.Fatalf("expected Parse section in result HTML: %s", truncate(resultHTML, 800))
	}
	if !strings.Contains(resultHTML, "Analyze") {
		t.Fatalf("expected Analyze section in result HTML: %s", truncate(resultHTML, 800))
	}
	if !strings.Contains(resultHTML, "QueryStatement") && !strings.Contains(resultHTML, "QueryStmt") {
		t.Fatalf("expected analyzer output marker in result HTML: %s", truncate(resultHTML, 800))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
