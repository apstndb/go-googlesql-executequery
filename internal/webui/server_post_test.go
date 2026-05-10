package webui

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandlerRunAcceptsMultipartFormData(t *testing.T) {
	t.Parallel()
	srv := NewServer(0)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField("sql", "SELECT 1 AS x"); err != nil {
		t.Fatal(err)
	}
	if err := mw.WriteField("catalog", "sample"); err != nil {
		t.Fatal(err)
	}
	if err := mw.WriteField("mode", "parse"); err != nil {
		t.Fatal(err)
	}
	if err := mw.WriteField("mode", "analyze"); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/run", &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, `class="result-section"`) {
		t.Fatalf("unexpected body: %s", truncateRunTest(s, 600))
	}
}

func TestHandlerRunAcceptsQueryFieldAlias(t *testing.T) {
	t.Parallel()
	srv := NewServer(0)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	form := url.Values{}
	form.Set("query", "SELECT 1 AS x")
	form.Set("catalog", "sample")
	form.Add("mode", "parse")

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/run", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `class="result-section"`) {
		t.Fatalf("unexpected body: %s", truncateRunTest(string(body), 600))
	}
}

func TestHandlerRunAcceptsURLEncodedForm(t *testing.T) {
	t.Parallel()
	srv := NewServer(0)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	form := url.Values{}
	form.Set("sql", "SELECT 1 AS x")
	form.Set("catalog", "sample")
	form.Add("mode", "parse")
	form.Add("mode", "analyze")

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/run", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, `class="result-section"`) {
		t.Fatalf("unexpected body: %s", truncateRunTest(s, 600))
	}
}

func truncateRunTest(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
