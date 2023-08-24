package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	mock_db "github.com/anshul393/snippetbox/mock"
	"github.com/go-playground/assert/v2"
	"github.com/go-playground/form/v4"
	"github.com/golang/mock/gomock"
)

type testServer struct {
	http.Server
	app *application
}

func TestSnippetCreate(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes(), app)
	go func() {
		ts.ListenAndServe()
	}()
	defer ts.Close()

	t.Run("Unauthenticated", func(t *testing.T) {
		code, headers, err := ts.get(t, "/snippet/create")
		assert.Equal(t, err, nil)
		assert.Equal(t, code,
			http.StatusSeeOther)
		assert.Equal(t, headers.Get("Location"), "/user/login")
	})

	t.Run("Authenticated", func(t *testing.T) {
		// Make a GET /user/login request and extract the CSRF token from the
		// response.
		_, _, body := ts.get(t, "/user/login")
		csrfToken := extractCSRFToken(t, body)
		// Make a POST /user/login request using the extracted CSRF token and
		// credentials from our the mock user model.
		form := url.Values{}
		form.Add("email", "alice@example.com")
		form.Add("password", "pa$$word")
		form.Add("csrf_token", csrfToken)
		ts.postForm(t, "/user/login", form)
		// Then check that the authenticated user is shown the create snippet
		// form.
		code, _, body := ts.get(t, "/snippet/create")
		bodydata, err := io.ReadAll(body)
		assert.Equal(t, err, nil)
		assert.Equal(t, code, http.StatusOK)
		assert.MatchRegex(t, string(bodydata), "*<form action='/snippet/create' method='POST'>*")
	})
}

func newTestApplication(t *testing.T) *application {
	tc, _ := newTemplateCache()
	return &application{
		dtb:           mock_db.NewMockQuerier(gomock.NewController(t)),
		templateCache: tc,
		formDecoder:   form.NewDecoder(),
	}
}

func newTestServer(t *testing.T, handler http.Handler, app *application) *testServer {
	return &testServer{Server: http.Server{
		Addr:    "localhost:5000",
		Handler: handler,
	}, app: app}
}

func (ts *testServer) get(t *testing.T, path string) (int, http.Header, io.ReadCloser) {
	resp, _ := http.Get("localhost:5000/snippet/create")
	return resp.StatusCode, resp.Header, resp.Body
}

func (ts *testServer) postForm(t *testing.T, path string, form url.Values) {
	http.PostForm(fmt.Sprintf("localhost:5000%s", path), form)
}

func extractCSRFToken(t *testing.T, body io.ReadCloser) string {
	return ""
}
