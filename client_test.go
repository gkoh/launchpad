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

		w.WriteHeader(StatusContentReturned)
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

	if resp.StatusCode != StatusContentReturned {
		t.Errorf("status = %d, want %d", resp.StatusCode, StatusContentReturned)
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

func TestBugSetTitle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		if r.URL.Path != "/bugs/100" {
			t.Errorf("path = %q, want /bugs/100", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"title":"New Title"`) {
			t.Errorf("body = %q, want title field", string(body))
		}
		w.WriteHeader(StatusContentReturned)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	bug := &Bug{client: client, ID: 100}
	if err := bug.SetTitle("New Title"); err != nil {
		t.Fatal(err)
	}
}

func TestBugGetTasks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"total_size": 1,
			"start": 0,
			"entries": [
				{
					"bug_target_name": "linux",
					"status": "Confirmed",
					"importance": "High",
					"self_link": "https://api.launchpad.net/devel/ubuntu/+source/linux/+bug/100"
				}
			]
		}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	bug := &Bug{
		client:                 client,
		ID:                     100,
		BugTasksCollectionLink: NewLink(srv.URL + "/bugs/100/tasks"),
	}

	tasks, err := bug.GetTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].BugTargetName != "linux" {
		t.Errorf("BugTargetName = %q, want %q", tasks[0].BugTargetName, "linux")
	}
	if tasks[0].Status != BugTaskStatusConfirmed {
		t.Errorf("Status = %q, want %q", tasks[0].Status, BugTaskStatusConfirmed)
	}
	if tasks[0].client == nil {
		t.Error("task client is nil, want non-nil")
	}
}

func TestBugTaskSetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"status":"In Progress"`) {
			t.Errorf("body = %q, want status field", string(body))
		}
		w.WriteHeader(StatusContentReturned)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)

	task := &BugTask{
		client:        client,
		BugTargetName: "linux",
		SelfLink:      NewLink(srv.URL + "/task/1"),
	}
	if err := task.SetStatus("In Progress"); err != nil {
		t.Fatal(err)
	}
}

func TestBugTaskSetImportance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"importance":"Critical"`) {
			t.Errorf("body = %q, want importance field", string(body))
		}
		w.WriteHeader(StatusContentReturned)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)

	task := &BugTask{
		client:        client,
		BugTargetName: "linux",
		SelfLink:      NewLink(srv.URL + "/task/1"),
	}
	if err := task.SetImportance("Critical"); err != nil {
		t.Fatal(err)
	}
}

func TestBugTaskSetAssignee(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"assignee_link":"https://api.launchpad.net/devel/~jdoe"`) {
			t.Errorf("body = %q, want assignee_link field", string(body))
		}
		w.WriteHeader(StatusContentReturned)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)

	task := &BugTask{
		client:        client,
		BugTargetName: "linux",
		SelfLink:      NewLink(srv.URL + "/task/1"),
	}
	if err := task.SetAssignee("jdoe"); err != nil {
		t.Fatal(err)
	}
}

func TestBugTaskSetAssigneeUnassign(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"assignee_link":null`) {
			t.Errorf("body = %q, want null assignee_link", string(body))
		}
		w.WriteHeader(StatusContentReturned)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)

	task := &BugTask{
		client:        client,
		BugTargetName: "linux",
		SelfLink:      NewLink(srv.URL + "/task/1"),
	}
	if err := task.SetAssignee(""); err != nil {
		t.Fatal(err)
	}
}

func TestBugGetMessages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"total_size": 2,
			"start": 0,
			"entries": [
				{"content": "first comment", "subject": "Bug #1", "owner_link": "https://api.launchpad.net/devel/~alice"},
				{"content": "second comment", "subject": "Re: Bug #1", "owner_link": "https://api.launchpad.net/devel/~bob"}
			]
		}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)

	bug := &Bug{
		client:                 client,
		ID:                     1,
		MessagesCollectionLink: NewLink(srv.URL + "/bugs/1/messages"),
	}

	messages, err := bug.GetMessages()
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(messages))
	}
	if messages[0].Content != "first comment" {
		t.Errorf("messages[0].Content = %q, want %q", messages[0].Content, "first comment")
	}
	if messages[1].Content != "second comment" {
		t.Errorf("messages[1].Content = %q, want %q", messages[1].Content, "second comment")
	}
}

func TestBugGetMessagesPaginated(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "" {
			w.Write([]byte(`{
				"total_size": 2,
				"start": 0,
				"next_collection_link": "` + srv.URL + `/bugs/1/messages?page=2",
				"entries": [
					{"content": "page one", "owner_link": "https://api.launchpad.net/devel/~alice"}
				]
			}`))
		} else {
			w.Write([]byte(`{
				"total_size": 2,
				"start": 1,
				"entries": [
					{"content": "page two", "owner_link": "https://api.launchpad.net/devel/~bob"}
				]
			}`))
		}
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)

	bug := &Bug{
		client:                 client,
		ID:                     1,
		MessagesCollectionLink: NewLink(srv.URL + "/bugs/1/messages"),
	}

	messages, err := bug.GetMessages()
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(messages))
	}
	if messages[0].Content != "page one" {
		t.Errorf("messages[0].Content = %q, want %q", messages[0].Content, "page one")
	}
	if messages[1].Content != "page two" {
		t.Errorf("messages[1].Content = %q, want %q", messages[1].Content, "page two")
	}
}

func TestResolveTaskAssignees(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name": "jdoe", "display_name": "John Doe"}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)

	tasks := []BugTask{
		{AssigneeLink: NewLink(srv.URL + "/~jdoe")},
		{}, // zero link, should be skipped
	}

	result := ResolveTaskAssignees(client, tasks)
	key := srv.URL + "/~jdoe"
	if result[key] != "John Doe (jdoe)" {
		t.Errorf("result[%q] = %q, want %q", key, result[key], "John Doe (jdoe)")
	}
	if len(result) != 1 {
		t.Errorf("got %d entries, want 1", len(result))
	}
}

func TestResolveMessageOwners(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name": "alice", "display_name": "Alice A"}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)

	messages := []Message{
		{OwnerLink: NewLink(srv.URL + "/~alice")},
		{OwnerLink: NewLink(srv.URL + "/~alice")}, // duplicate, should dedupe
	}

	result := ResolveMessageOwners(client, messages)
	key := srv.URL + "/~alice"
	if result[key] != "Alice A (alice)" {
		t.Errorf("result[%q] = %q, want %q", key, result[key], "Alice A (alice)")
	}
	if len(result) != 1 {
		t.Errorf("got %d entries, want 1", len(result))
	}
}
