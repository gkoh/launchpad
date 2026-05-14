package launchpad

import (
	"encoding/json"
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

// Me fetches the authenticated user's details from /people/+me and
// returns the parsed Person.
func (c *Client) Me() (*Person, error) {
	resp, err := c.Get("/people/+me")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("launchpad: reading me response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("launchpad: me returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var person Person
	if err := json.Unmarshal(body, &person); err != nil {
		return nil, fmt.Errorf("launchpad: parsing me response: %w", err)
	}

	return &person, nil
}

// GetPerson fetches a Person by their full API link URL.
func (c *Client) GetPerson(link Link) (*Person, error) {
	if link.IsZero() {
		return nil, fmt.Errorf("launchpad: empty person link")
	}

	req, err := http.NewRequest(http.MethodGet, link.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("launchpad: reading person response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("launchpad: person returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var person Person
	if err := json.Unmarshal(body, &person); err != nil {
		return nil, fmt.Errorf("launchpad: parsing person response: %w", err)
	}

	return &person, nil
}

// ResolvePersonLinks fetches the display name for each unique non-zero Link.
// Returns a map from link URL string to "DisplayName (name)".
// Links that fail to resolve are mapped to their raw URL string.
func (c *Client) ResolvePersonLinks(links []Link) map[string]string {
	result := make(map[string]string)

	for _, link := range links {
		if link.IsZero() {
			continue
		}
		url := link.String()
		if _, ok := result[url]; ok {
			continue
		}

		person, err := c.GetPerson(link)
		if err != nil {
			result[url] = url
			continue
		}

		result[url] = fmt.Sprintf("%s (%s)", person.DisplayName, person.Name)
	}

	return result
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
