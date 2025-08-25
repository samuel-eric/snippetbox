package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/samuel-eric/snippetbox/internal/models/mocks"
)

func newTestApplication(t *testing.T) *application {
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &application{
		logger:         slog.New(slog.DiscardHandler),
		snippets:       &mocks.SnippetModel{},
		users:          &mocks.UserModel{},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

type testResponse struct {
	status  int
	headers http.Header
	cookies []*http.Cookie
	body    string
}

func (ts *testServer) get(t *testing.T, urlPath string) testResponse {
	req, err := http.NewRequest(http.MethodGet, ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	return testResponse{
		status:  res.StatusCode,
		headers: res.Header,
		cookies: res.Cookies(),
		body:    string(bytes.TrimSpace(body)),
	}
}

func (ts *testServer) resetClientCookieJar(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	ts.Client().Jar = jar
}
