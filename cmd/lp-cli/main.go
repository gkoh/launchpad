// Command lp-cli performs the interactive OAuth 1.0a flow to obtain
// Launchpad API credentials, saves them to disk, and verifies them
// by fetching the authenticated user's profile.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gkoh/launchpad"
)

func main() {
	consumerKey := flag.String("consumer", "lp-cli", "OAuth consumer key (application name)")
	permission := flag.String("permission", launchpad.PermissionReadPrivate, "Permission level (READ_PUBLIC, READ_PRIVATE, WRITE_PUBLIC, WRITE_PRIVATE)")
	credsPath := flag.String("credentials", "", "Path to save credentials (default: ~/.config/launchpad/<consumer>.json)")
	check := flag.Bool("check", false, "Check that stored credentials are valid and exit")
	flag.Parse()

	if *credsPath == "" {
		p, err := launchpad.DefaultCredentialsPath(*consumerKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		*credsPath = p
	}

	// -check: load and verify stored credentials, then exit.
	if *check {
		creds, err := launchpad.LoadCredentials(*credsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading credentials from %s: %v\n", *credsPath, err)
			os.Exit(1)
		}
		if err := verify(creds); err != nil {
			fmt.Fprintf(os.Stderr, "Credentials invalid: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check for existing credentials.
	if creds, err := launchpad.LoadCredentials(*credsPath); err == nil {
		fmt.Printf("Found existing credentials at %s\n", *credsPath)
		fmt.Println("Verifying...")
		if err := verify(creds); err != nil {
			fmt.Fprintf(os.Stderr, "Existing credentials are invalid: %v\n", err)
			fmt.Println("Re-authenticating...")
		} else {
			return
		}
	}

	cfg := &launchpad.AuthConfig{
		ConsumerKey: *consumerKey,
	}

	// Step 1: Get request token.
	fmt.Println("Requesting token from Launchpad...")
	reqToken, err := launchpad.GetRequestToken(cfg, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Step 2: User authorizes.
	authURL := launchpad.AuthorizeURL(cfg, reqToken, *permission)
	fmt.Println()
	fmt.Println("Please open the following URL in your browser and authorize the application:")
	fmt.Println()
	fmt.Printf("  %s\n", authURL)
	fmt.Println()
	fmt.Print("Press Enter after you have authorized the application...")
	bufio.NewReader(os.Stdin).ReadLine()

	// Step 3: Exchange for access token.
	fmt.Println("Exchanging for access token...")
	accessToken, err := launchpad.ExchangeRequestToken(cfg, reqToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	creds := &launchpad.Credentials{
		ConsumerKey: *consumerKey,
		Token:       accessToken,
	}

	// Save credentials.
	if err := creds.Save(*credsPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving credentials: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Credentials saved to %s\n", *credsPath)

	// Verify.
	fmt.Println("Verifying credentials...")
	if err := verify(creds); err != nil {
		fmt.Fprintf(os.Stderr, "Verification failed: %v\n", err)
		os.Exit(1)
	}
}

func verify(creds *launchpad.Credentials) error {
	client := launchpad.NewClient(creds, nil)
	resp, err := client.Me()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var user struct {
		DisplayName string `json:"display_name"`
		Name        string `json:"name"`
		WebLink     string `json:"web_link"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fmt.Printf("Authenticated as: %s (%s)\n", user.DisplayName, user.Name)
	fmt.Printf("Profile: %s\n", user.WebLink)
	return nil
}
