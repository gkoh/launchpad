package launchpad

import (
	"io"
	"net/http"
	"net/http/httptest"
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
