package launchpad

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is an HTTP client that automatically signs requests to the
// Launchpad API using OAuth 1.0a PLAINTEXT credentials.
type Client struct {
	// APIBaseURL is the root of the Launchpad API.
	// Defaults to DefaultAPIBaseURL if empty.
	APIBaseURL string

	creds      *Credentials
	httpClient *http.Client
}

// NewClient returns a Client that signs every request with the given credentials.
// An optional base http.Client can be provided; if nil, http.DefaultClient is used.
func NewClient(creds *Credentials, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &oauthTransport{
				base:  http.DefaultTransport,
				creds: creds,
			},
		}
	} else {
		base := httpClient.Transport
		if base == nil {
			base = http.DefaultTransport
		}
		httpClient.Transport = &oauthTransport{
			base:  base,
			creds: creds,
		}
	}

	return &Client{
		creds:      creds,
		httpClient: httpClient,
	}
}

func (c *Client) apiBase() string {
	if c.APIBaseURL != "" {
		return strings.TrimRight(c.APIBaseURL, "/")
	}
	return DefaultAPIBaseURL
}

// Do executes an arbitrary HTTP request, signing it with OAuth credentials.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

// Get performs a GET request against the Launchpad API.
// The path is relative to the API base URL (e.g. "/+me" or "/bugs/12345").
func (c *Client) Get(path string) (*http.Response, error) {
	u := c.apiBase() + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("launchpad: creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	return c.Do(req)
}

// Me is a convenience method that fetches the authenticated user's details
// from /people/+me on the Launchpad API.
func (c *Client) Me() (*http.Response, error) {
	return c.Get("/people/+me")
}

// Patch performs a PATCH request against the Launchpad API with a JSON body.
// The path is relative to the API base URL (e.g. "/bugs/12345").
func (c *Client) Patch(path string, body io.Reader) (*http.Response, error) {
	u := c.apiBase() + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequest(http.MethodPatch, u, body)
	if err != nil {
		return nil, fmt.Errorf("launchpad: creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.Do(req)
}

// oauthTransport is an http.RoundTripper that injects OAuth headers.
type oauthTransport struct {
	base  http.RoundTripper
	creds *Credentials
}

func (t *oauthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone to avoid mutating the original request.
	r := req.Clone(req.Context())
	SignRequest(r, t.creds.ConsumerKey, t.creds.Token)
	return t.base.RoundTrip(r)
}
