package launchpad_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gkoh/launchpad"
)

// Example_fullFlow demonstrates the complete OAuth 1.0a authentication flow
// against a mock Launchpad server.
func Example_fullFlow() {
	// Set up a mock Launchpad server.
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/+request-token":
			w.Write([]byte("oauth_token=reqtok&oauth_token_secret=reqsec"))
		case "/+access-token":
			w.Write([]byte("oauth_token=acctok&oauth_token_secret=accsec"))
		default:
			// API endpoint
			w.Write([]byte(`{"display_name":"Example User","name":"example"}`))
		}
	}))
	defer mock.Close()

	cfg := &launchpad.AuthConfig{
		ConsumerKey: "example-app",
		BaseURL:     mock.URL,
	}

	// Step 1: Get request token.
	reqToken, err := launchpad.GetRequestToken(cfg, "")
	if err != nil {
		panic(err)
	}

	// Step 2: Build the authorization URL (user would visit this).
	authURL := launchpad.AuthorizeURL(cfg, reqToken, launchpad.PermissionReadPublic)
	_ = authURL // In a real app, open this URL in the user's browser.

	// Step 3: Exchange for access token (after user approves).
	accessToken, err := launchpad.ExchangeRequestToken(cfg, reqToken)
	if err != nil {
		panic(err)
	}

	// Step 4: Use the authenticated client.
	creds := &launchpad.Credentials{
		ConsumerKey: cfg.ConsumerKey,
		Token:       accessToken,
	}
	client := launchpad.NewClient(creds, nil)
	client.APIBaseURL = mock.URL

	resp, err := client.Get("/+me")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
	// Output: Status: 200
}
