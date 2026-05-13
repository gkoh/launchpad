// Command lp-cli performs the interactive OAuth 1.0a flow to obtain
// Launchpad API credentials, saves them to disk, and verifies them
// by fetching the authenticated user's profile.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gkoh/launchpad"
)

func main() {
	consumerKey := flag.String("consumer", "lp-cli", "OAuth consumer key (application name)")
	permission := flag.String("permission", launchpad.PermissionReadPrivate, "Permission level (READ_PUBLIC, READ_PRIVATE, WRITE_PUBLIC, WRITE_PRIVATE)")
	credsPath := flag.String("credentials", "", "Path to save credentials (default: ~/.config/launchpad/<consumer>.json)")
	check := flag.Bool("check", false, "Check that stored credentials are valid and exit")
	bugID := flag.Int("bug", 0, "Query and display a bug by its ID")
	setTitle := flag.String("set-title", "", "Set the bug title (requires -bug)")
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

	// -bug: query and display a bug by ID; optionally update its title.
	if *bugID > 0 {
		creds, err := launchpad.LoadCredentials(*credsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading credentials from %s: %v\n", *credsPath, err)
			os.Exit(1)
		}
		client := launchpad.NewClient(creds, nil)

		if *setTitle != "" {
			if err := updateBugTitle(client, *bugID, *setTitle); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		if err := showBug(client, *bugID); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *setTitle != "" {
		fmt.Fprintf(os.Stderr, "Error: -set-title requires -bug\n")
		os.Exit(1)
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

func showBug(client *launchpad.Client, bugID int) error {

	// Fetch the bug.
	resp, err := client.Get(fmt.Sprintf("/bugs/%d", bugID))
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

	var bug launchpad.Bug
	if err := json.Unmarshal(body, &bug); err != nil {
		return fmt.Errorf("parsing bug: %w", err)
	}

	// Display bug summary.
	fmt.Printf("Bug #%d: %s\n", bug.ID, bug.Title)
	fmt.Printf("Info type:   %s\n", bug.InformationType)
	fmt.Printf("Lock status: %s\n", bug.LockStatus)
	fmt.Printf("Heat:        %d\n", bug.Heat)
	if len(bug.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(bug.Tags, ", "))
	}
	if bug.DateCreated != nil {
		fmt.Printf("Created:     %s\n", bug.DateCreated.Format("2006-01-02 15:04:05"))
	}
	if bug.DateLastUpdated != nil {
		fmt.Printf("Updated:     %s\n", bug.DateLastUpdated.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("Messages:    %d\n", bug.MessageCount)
	if bug.Description != "" {
		desc := bug.Description
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		fmt.Printf("Description: %s\n", desc)
	}

	// Fetch bug tasks.
	if bug.BugTasksCollectionLink != "" {
		tasks, err := fetchBugTasks(client, bug.BugTasksCollectionLink)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nWarning: could not fetch bug tasks: %v\n", err)
		} else if len(tasks) > 0 {
			// Resolve assignees.
			assignees := resolveAssignees(client, tasks)

			fmt.Println()
			fmt.Println("Tasks:")
			for _, task := range tasks {
				fmt.Printf("  %s\n", task.BugTargetDisplayName)
				fmt.Printf("    Status:     %s\n", task.Status)
				fmt.Printf("    Importance: %s\n", task.Importance)
				if name, ok := assignees[task.AssigneeLink]; ok {
					fmt.Printf("    Assignee:   %s\n", name)
				} else {
					fmt.Printf("    Assignee:   unassigned\n")
				}
			}
		}
	}

	fmt.Printf("\nWeb:         %s\n", bug.WebLink)

	return nil
}

// fetchBugTasks fetches the bug task collection from the given URL.
func fetchBugTasks(client *launchpad.Client, url string) ([]launchpad.BugTask, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bug tasks returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var collection launchpad.BugTaskCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, err
	}

	return collection.Entries, nil
}

// resolveAssignees fetches the display name for each unique assignee link.
// Returns a map from assignee_link URL to "DisplayName (name)" string.
func resolveAssignees(client *launchpad.Client, tasks []launchpad.BugTask) map[string]string {
	result := make(map[string]string)

	for _, task := range tasks {
		if task.AssigneeLink == "" {
			continue
		}
		if _, ok := result[task.AssigneeLink]; ok {
			continue // already resolved
		}

		req, err := http.NewRequest(http.MethodGet, task.AssigneeLink, nil)
		if err != nil {
			result[task.AssigneeLink] = task.AssigneeLink
			continue
		}
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			result[task.AssigneeLink] = task.AssigneeLink
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != 200 {
			result[task.AssigneeLink] = task.AssigneeLink
			continue
		}

		var person launchpad.Person
		if err := json.Unmarshal(body, &person); err != nil {
			result[task.AssigneeLink] = task.AssigneeLink
			continue
		}

		result[task.AssigneeLink] = fmt.Sprintf("%s (%s)", person.DisplayName, person.Name)
	}

	return result
}

// updateBugTitle sends a PATCH request to update a bug's title.
func updateBugTitle(client *launchpad.Client, bugID int, title string) error {
	payload, err := json.Marshal(map[string]string{"title": title})
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}

	resp, err := client.Patch(fmt.Sprintf("/bugs/%d", bugID), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("updating bug title: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 209 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	fmt.Printf("Title updated to: %s\n\n", title)
	return nil
}
