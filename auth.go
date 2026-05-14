// Package launchpad provides OAuth 1.0a authentication for Canonical's
// Launchpad web service API (https://api.launchpad.net).
//
// Launchpad uses OAuth 1.0a with PLAINTEXT signing. The authentication flow is:
//
//  1. Obtain a request token from https://launchpad.net/+request-token
//  2. Direct the user to https://launchpad.net/+authorize-token to approve access
//  3. Exchange the authorized request token for an access token at https://launchpad.net/+access-token
//  4. Sign subsequent API requests with the access token
package launchpad

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// Launchpad OAuth endpoints.
	defaultBaseURL     = "https://launchpad.net"
	requestTokenPath   = "/+request-token"
	authorizeTokenPath = "/+authorize-token"
	accessTokenPath    = "/+access-token"

	// DefaultAPIBaseURL is the base URL for the Launchpad devel API.
	DefaultAPIBaseURL = "https://api.launchpad.net/devel"

	// OAuth signature method used by Launchpad.
	signatureMethod = "PLAINTEXT"
)

// Permission levels for Launchpad API access.
const (
	PermissionReadPublic   = "READ_PUBLIC"
	PermissionReadPrivate  = "READ_PRIVATE"
	PermissionWritePublic  = "WRITE_PUBLIC"
	PermissionWritePrivate = "WRITE_PRIVATE"
)

// RequestToken holds the temporary OAuth request token returned by Launchpad.
type RequestToken struct {
	Token  string
	Secret string
}

// AccessToken holds the OAuth access token and secret used to sign API requests.
type AccessToken struct {
	Token  string
	Secret string
}

// AuthConfig holds configuration for the OAuth flow.
type AuthConfig struct {
	// ConsumerKey identifies the application requesting access.
	ConsumerKey string

	// BaseURL is the Launchpad instance base URL.
	// Defaults to https://launchpad.net if empty.
	BaseURL string

	// HTTPClient is the HTTP client used during the OAuth flow.
	// Defaults to http.DefaultClient if nil.
	HTTPClient *http.Client
}

func (c *AuthConfig) baseURL() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	return defaultBaseURL
}

func (c *AuthConfig) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

// oauthSignature computes the PLAINTEXT OAuth signature.
// For Launchpad the consumer secret is always empty, so the signature
// is "&<token_secret>".
func oauthSignature(tokenSecret string) string {
	return "&" + tokenSecret
}

// GetRequestToken obtains an OAuth request token from Launchpad.
// The context parameter is an optional string displayed to the user during
// authorization (e.g. "my-app wants to access Launchpad on your behalf").
func GetRequestToken(cfg *AuthConfig, context string) (*RequestToken, error) {
	data := url.Values{
		"oauth_consumer_key":     {cfg.ConsumerKey},
		"oauth_signature_method": {signatureMethod},
		"oauth_signature":        {oauthSignature("")},
	}
	if context != "" {
		data.Set("lp.context", context)
	}

	endpoint := cfg.baseURL() + requestTokenPath
	resp, err := cfg.httpClient().PostForm(endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("launchpad: request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("launchpad: reading request token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("launchpad: request token returned %d: %s", resp.StatusCode, string(body))
	}

	vals, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("launchpad: parsing request token response: %w", err)
	}

	return &RequestToken{
		Token:  vals.Get("oauth_token"),
		Secret: vals.Get("oauth_token_secret"),
	}, nil
}

// AuthorizeURL returns the URL the user must visit to authorize the request token.
// permission should be one of the Permission* constants.
func AuthorizeURL(cfg *AuthConfig, token *RequestToken, permission string) string {
	params := url.Values{
		"oauth_token":      {token.Token},
		"allow_permission": {permission},
	}
	return cfg.baseURL() + authorizeTokenPath + "?" + params.Encode()
}

// ExchangeRequestToken exchanges an authorized request token for an access token.
func ExchangeRequestToken(cfg *AuthConfig, reqToken *RequestToken) (*AccessToken, error) {
	data := url.Values{
		"oauth_consumer_key":     {cfg.ConsumerKey},
		"oauth_token":            {reqToken.Token},
		"oauth_signature_method": {signatureMethod},
		"oauth_signature":        {oauthSignature(reqToken.Secret)},
	}

	endpoint := cfg.baseURL() + accessTokenPath
	resp, err := cfg.httpClient().PostForm(endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("launchpad: exchange token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("launchpad: reading access token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("launchpad: exchange token returned %d: %s", resp.StatusCode, string(body))
	}

	vals, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("launchpad: parsing access token response: %w", err)
	}

	return &AccessToken{
		Token:  vals.Get("oauth_token"),
		Secret: vals.Get("oauth_token_secret"),
	}, nil
}

// SignRequest adds OAuth 1.0a PLAINTEXT authorization headers to an HTTP request.
func SignRequest(req *http.Request, consumerKey string, token *AccessToken) {
	sig := url.QueryEscape(oauthSignature(token.Secret))
	authHeader := fmt.Sprintf(
		`OAuth realm="https://api.launchpad.net/", `+
			`oauth_consumer_key="%s", `+
			`oauth_token="%s", `+
			`oauth_signature_method="%s", `+
			`oauth_signature="%s", `+
			`oauth_version="1.0"`,
		url.QueryEscape(consumerKey),
		url.QueryEscape(token.Token),
		signatureMethod,
		sig,
	)
	req.Header.Set("Authorization", authHeader)
}
