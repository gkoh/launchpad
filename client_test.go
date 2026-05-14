package launchpad

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// statusContentReturned is Launchpad's non-standard HTTP 209 response
// indicating a successful PATCH with content returned.
const statusContentReturned = 209

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
		w.Write([]byte(`{"display_name":"Test User","name":"testuser","web_link":"https://launchpad.net/~testuser"}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	person, err := client.Me()
	if err != nil {
		t.Fatal(err)
	}

	if person.DisplayName != "Test User" {
		t.Errorf("DisplayName = %q, want %q", person.DisplayName, "Test User")
	}
	if person.Name != "testuser" {
		t.Errorf("Name = %q, want %q", person.Name, "testuser")
	}
}

func TestOAuthTransportSigns(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
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

		w.WriteHeader(statusContentReturned)
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

	if resp.StatusCode != statusContentReturned {
		t.Errorf("status = %d, want %d", resp.StatusCode, statusContentReturned)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"title":"new title"}` {
		t.Errorf("response body = %q", string(body))
	}
}

func TestClientGetBug(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bugs/12345" {
			t.Errorf("path = %q, want /bugs/12345", r.URL.Path)
		}
		w.Write([]byte(`{
			"id": 12345,
			"title": "Test bug title",
			"description": "A test bug",
			"heat": 42,
			"tags": ["kernel", "sru"],
			"information_type": "Public",
			"message_count": 3,
			"web_link": "https://bugs.launchpad.net/ubuntu/+bug/12345",
			"self_link": "https://api.launchpad.net/devel/bugs/12345"
		}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	bug, err := client.GetBug(12345)
	if err != nil {
		t.Fatal(err)
	}

	if bug.ID != 12345 {
		t.Errorf("ID = %d, want 12345", bug.ID)
	}
	if bug.Title != "Test bug title" {
		t.Errorf("Title = %q, want %q", bug.Title, "Test bug title")
	}
	if bug.Description != "A test bug" {
		t.Errorf("Description = %q, want %q", bug.Description, "A test bug")
	}
	if bug.Heat != 42 {
		t.Errorf("Heat = %d, want 42", bug.Heat)
	}
	if len(bug.Tags) != 2 || bug.Tags[0] != "kernel" || bug.Tags[1] != "sru" {
		t.Errorf("Tags = %v, want [kernel sru]", bug.Tags)
	}
	if bug.InformationType != InformationPublic {
		t.Errorf("InformationType = %q, want %q", bug.InformationType, InformationPublic)
	}
	if bug.MessageCount != 3 {
		t.Errorf("MessageCount = %d, want 3", bug.MessageCount)
	}
}

func TestClientGetBugNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "bug not found"}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	_, err := client.GetBug(99999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %q, want it to contain 404", err.Error())
	}
}
