package launchpad

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTagsCombinatorJSON(t *testing.T) {
	tests := []struct {
		val  TagsCombinator
		want string
	}{
		{TagsCombinatorAll, `"All"`},
		{TagsCombinatorAny, `"Any"`},
	}

	for _, tt := range tests {
		data, err := json.Marshal(tt.val)
		if err != nil {
			t.Errorf("Marshal(%q): %v", tt.val, err)
			continue
		}
		if string(data) != tt.want {
			t.Errorf("Marshal(%q) = %s, want %s", tt.val, data, tt.want)
		}
	}
}

func TestSearchTasksBasic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ubuntu" {
			t.Errorf("path = %q, want /ubuntu", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("ws.op") != "searchTasks" {
			t.Errorf("ws.op = %q, want searchTasks", q.Get("ws.op"))
		}
		if q.Get("search_text") != "kernel" {
			t.Errorf("search_text = %q, want kernel", q.Get("search_text"))
		}
		if q.Get("status") != "Confirmed" {
			t.Errorf("status = %q, want Confirmed", q.Get("status"))
		}
		if q.Get("importance") != "High" {
			t.Errorf("importance = %q, want High", q.Get("importance"))
		}
		if !strings.HasSuffix(q.Get("assignee"), "/~jdoe") {
			t.Errorf("assignee = %q, want suffix /~jdoe", q.Get("assignee"))
		}
		if q.Get("ws.size") != "10" {
			t.Errorf("ws.size = %q, want 10", q.Get("ws.size"))
		}

		w.Write([]byte(`{
			"total_size": 1,
			"start": 0,
			"entries": [
				{
					"bug_target_name": "linux",
					"status": "Confirmed",
					"importance": "High",
					"title": "Test task",
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

	tasks, err := client.SearchTasks("ubuntu", &SearchTasksOptions{
		SearchText: "kernel",
		Status:     "Confirmed",
		Importance: "High",
		Assignee:   "jdoe",
		PageSize:   10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].Title != "Test task" {
		t.Errorf("Title = %q, want %q", tasks[0].Title, "Test task")
	}
	if tasks[0].client == nil {
		t.Error("task client is nil, want non-nil")
	}
}

func TestSearchTasksNilOpts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("ws.op") != "searchTasks" {
			t.Errorf("ws.op = %q, want searchTasks", q.Get("ws.op"))
		}
		w.Write([]byte(`{"total_size": 0, "start": 0, "entries": []}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	tasks, err := client.SearchTasks("ubuntu", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("got %d tasks, want 0", len(tasks))
	}
}

func TestSearchTasksTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		tags := q["tags"]
		if len(tags) != 2 || tags[0] != "kernel" || tags[1] != "sru" {
			t.Errorf("tags = %v, want [kernel sru]", tags)
		}
		if q.Get("tags_combinator") != "Any" {
			t.Errorf("tags_combinator = %q, want Any", q.Get("tags_combinator"))
		}
		w.Write([]byte(`{"total_size": 0, "start": 0, "entries": []}`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	_, err := client.SearchTasks("ubuntu", &SearchTasksOptions{
		Tags:           []string{"kernel", "sru"},
		TagsCombinator: TagsCombinatorAny,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchTasksPaginated(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "" {
			w.Write([]byte(`{
				"total_size": 2,
				"start": 0,
				"next_collection_link": "` + srv.URL + `/ubuntu?page=2",
				"entries": [
					{"bug_target_name": "linux", "status": "New", "title": "first"}
				]
			}`))
		} else {
			w.Write([]byte(`{
				"total_size": 2,
				"start": 1,
				"entries": [
					{"bug_target_name": "linux", "status": "New", "title": "second"}
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
	client.APIBaseURL = srv.URL

	tasks, err := client.SearchTasks("ubuntu", &SearchTasksOptions{
		FollowPages: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
	if tasks[0].Title != "first" {
		t.Errorf("tasks[0].Title = %q, want %q", tasks[0].Title, "first")
	}
	if tasks[1].Title != "second" {
		t.Errorf("tasks[1].Title = %q, want %q", tasks[1].Title, "second")
	}
}

func TestSearchTasksNoPagination(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "" {
			t.Error("unexpected second page request when FollowPages is false")
		}
		w.Write([]byte(`{
			"total_size": 2,
			"start": 0,
			"next_collection_link": "` + srv.URL + `/ubuntu?page=2",
			"entries": [
				{"bug_target_name": "linux", "status": "New", "title": "first"}
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

	tasks, err := client.SearchTasks("ubuntu", &SearchTasksOptions{
		FollowPages: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
}

func TestSearchTasksNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`not found`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	_, err := client.SearchTasks("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want it to contain 'not found'", err.Error())
	}
}

func TestSearchTasksAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error`))
	}))
	defer srv.Close()

	creds := &Credentials{
		ConsumerKey: "test",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	client := NewClient(creds, nil)
	client.APIBaseURL = srv.URL

	_, err := client.SearchTasks("ubuntu", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want it to contain 500", err.Error())
	}
}
