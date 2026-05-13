package launchpad

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetRequestToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != requestTokenPath {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := r.FormValue("oauth_consumer_key"); got != "test-app" {
			t.Fatalf("consumer_key = %q, want %q", got, "test-app")
		}
		if got := r.FormValue("oauth_signature_method"); got != "PLAINTEXT" {
			t.Fatalf("signature_method = %q, want %q", got, "PLAINTEXT")
		}
		if got := r.FormValue("oauth_signature"); got != "&" {
			t.Fatalf("signature = %q, want %q", got, "&")
		}
		w.Write([]byte("oauth_token=req123&oauth_token_secret=sec456"))
	}))
	defer srv.Close()

	cfg := &AuthConfig{
		ConsumerKey: "test-app",
		BaseURL:     srv.URL,
	}

	tok, err := GetRequestToken(cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if tok.Token != "req123" {
		t.Errorf("Token = %q, want %q", tok.Token, "req123")
	}
	if tok.Secret != "sec456" {
		t.Errorf("Secret = %q, want %q", tok.Secret, "sec456")
	}
}

func TestGetRequestTokenWithContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := r.FormValue("lp.context"); got != "my context" {
			t.Fatalf("lp.context = %q, want %q", got, "my context")
		}
		w.Write([]byte("oauth_token=t&oauth_token_secret=s"))
	}))
	defer srv.Close()

	cfg := &AuthConfig{ConsumerKey: "test", BaseURL: srv.URL}
	_, err := GetRequestToken(cfg, "my context")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRequestTokenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	cfg := &AuthConfig{ConsumerKey: "test", BaseURL: srv.URL}
	_, err := GetRequestToken(cfg, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthorizeURL(t *testing.T) {
	cfg := &AuthConfig{ConsumerKey: "test", BaseURL: "https://launchpad.test"}
	tok := &RequestToken{Token: "mytoken", Secret: "mysecret"}
	got := AuthorizeURL(cfg, tok, PermissionReadPublic)

	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "launchpad.test" {
		t.Errorf("host = %q", u.Host)
	}
	if u.Path != authorizeTokenPath {
		t.Errorf("path = %q", u.Path)
	}
	if u.Query().Get("oauth_token") != "mytoken" {
		t.Errorf("oauth_token = %q", u.Query().Get("oauth_token"))
	}
	if u.Query().Get("allow_permission") != PermissionReadPublic {
		t.Errorf("allow_permission = %q", u.Query().Get("allow_permission"))
	}
}

func TestExchangeRequestToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != accessTokenPath {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := r.FormValue("oauth_token"); got != "reqtok" {
			t.Fatalf("oauth_token = %q", got)
		}
		if got := r.FormValue("oauth_signature"); got != "&reqsec" {
			t.Fatalf("signature = %q, want %q", got, "&reqsec")
		}
		w.Write([]byte("oauth_token=access123&oauth_token_secret=accsec456"))
	}))
	defer srv.Close()

	cfg := &AuthConfig{ConsumerKey: "test", BaseURL: srv.URL}
	reqTok := &RequestToken{Token: "reqtok", Secret: "reqsec"}

	acc, err := ExchangeRequestToken(cfg, reqTok)
	if err != nil {
		t.Fatal(err)
	}
	if acc.Token != "access123" {
		t.Errorf("Token = %q", acc.Token)
	}
	if acc.Secret != "accsec456" {
		t.Errorf("Secret = %q", acc.Secret)
	}
}

func TestSignRequest(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://api.launchpad.net/devel/+me", nil)
	tok := &AccessToken{Token: "tok", Secret: "sec"}
	SignRequest(req, "myapp", tok)

	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Fatal("Authorization header not set")
	}
	for _, want := range []string{
		`oauth_consumer_key="myapp"`,
		`oauth_token="tok"`,
		`oauth_signature_method="PLAINTEXT"`,
	} {
		if !strings.Contains(auth, want) {
			t.Errorf("Authorization header missing %q\ngot: %s", want, auth)
		}
	}
}

func TestOAuthSignature(t *testing.T) {
	if got := oauthSignature(""); got != "&" {
		t.Errorf("oauthSignature('') = %q, want %q", got, "&")
	}
	if got := oauthSignature("secret"); got != "&secret" {
		t.Errorf("oauthSignature('secret') = %q, want %q", got, "&secret")
	}
}
