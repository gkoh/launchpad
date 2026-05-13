package launchpad

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("missing Authorization header")
		}
		if r.URL.Path != "/+me" {
			t.Errorf("path = %q, want /+me", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		w.Write([]byte(`{"name":"testuser"}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	resp, err := client.Get("/+me")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"name":"testuser"}` {
		t.Errorf("body = %q", string(body))
	}
}

func TestClientMe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/people/+me" {
			t.Errorf("path = %q, want /people/+me", r.URL.Path)
		}
		w.Write([]byte(`{"display_name":"Test User"}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	resp, err := client.Me()
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

func TestOAuthTransportSigns(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "app",
		Token:       &AccessToken{Token: "tok", Secret: "sec"},
	}
	client := NewClient(creds, nil)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	client.Do(req)

	if gotAuth == "" {
		t.Error("request was not signed")
	}
}

func TestClientPatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		if r.URL.Path != "/bugs/42" {
			t.Errorf("path = %q, want /bugs/42", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}

		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"title":"new title"}` {
			t.Errorf("body = %q", string(body))
		}

		w.WriteHeader(209)
		w.Write([]byte(`{"title":"new title"}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	resp, err := client.Patch("/bugs/42", strings.NewReader(`{"title":"new title"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 209 {
		t.Errorf("status = %d, want 209", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"title":"new title"}` {
		t.Errorf("response body = %q", string(body))
	}
}
