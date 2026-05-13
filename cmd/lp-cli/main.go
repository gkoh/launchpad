// Command lp-cli is a CLI tool for interacting with the Launchpad API.
//
// Usage:
//
//	lp-cli [global flags] <command> [<resource>] [flags]
//
// Global flags:
//
//	-consumer      OAuth consumer key (default: "lp-cli")
//	-credentials   Path to credentials file
//
// Commands:
//
//	auth                         Authenticate with Launchpad
//	  -check                     Verify stored credentials
//	  -permission                Permission level (default: READ_PRIVATE)
//
//	get <resource> [flags]       Fetch and display a resource
//	  bug -id <int>              Query and display a bug
//
//	set <resource> [flags]       Update a resource
//	  bug -id <int> -title <s>   Update bug fields
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
	globalFlags := flag.NewFlagSet("lp-cli", flag.ContinueOnError)
	consumerKey := globalFlags.String("consumer", "lp-cli", "OAuth consumer key (application name)")
	credsPath := globalFlags.String("credentials", "", "Path to credentials file (default: ~/.config/launchpad/<consumer>.json)")

	globalFlags.Usage = usage

	if err := globalFlags.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	if *credsPath == "" {
		p, err := launchpad.DefaultCredentialsPath(*consumerKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		*credsPath = p
	}

	args := globalFlags.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	switch args[0] {
	case "auth":
		cmdAuth(*credsPath, *consumerKey, args[1:])
	case "get":
		cmdGet(*credsPath, *consumerKey, args[1:])
	case "set":
		cmdSet(*credsPath, *consumerKey, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: lp-cli [global flags] <command> [<resource>] [flags]

Global flags:
  -consumer      OAuth consumer key (default: "lp-cli")
  -credentials   Path to credentials file

Commands:
  auth                           Authenticate with Launchpad
    -check                       Verify stored credentials and exit
    -permission <level>          Permission level (default: READ_PRIVATE)

  get <resource> [flags]         Fetch and display a resource
    bug -id <int>                Query and display a bug by ID

  set <resource> [flags]         Update a resource
    bug -id <int> [-title <s>]   Update bug fields
`)
}

// cmdAuth handles the "auth" subcommand.
func cmdAuth(credsPath, consumerKey string, args []string) {
	fs := flag.NewFlagSet("auth", flag.ExitOnError)
	check := fs.Bool("check", false, "Verify stored credentials and exit")
	permission := fs.String("permission", launchpad.PermissionReadPrivate, "Permission level (READ_PUBLIC, READ_PRIVATE, WRITE_PUBLIC, WRITE_PRIVATE)")
	fs.Parse(args)

	if *check {
		creds, err := launchpad.LoadCredentials(credsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading credentials from %s: %v\n", credsPath, err)
			os.Exit(1)
		}
		if err := verify(creds); err != nil {
			fmt.Fprintf(os.Stderr, "Credentials invalid: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check for existing valid credentials.
	if creds, err := launchpad.LoadCredentials(credsPath); err == nil {
		fmt.Printf("Found existing credentials at %s\n", credsPath)
		fmt.Println("Verifying...")
		if err := verify(creds); err != nil {
			fmt.Fprintf(os.Stderr, "Existing credentials are invalid: %v\n", err)
			fmt.Println("Re-authenticating...")
		} else {
			return
		}
	}

	cfg := &launchpad.AuthConfig{
		ConsumerKey: consumerKey,
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
		ConsumerKey: consumerKey,
		Token:       accessToken,
	}

	// Save credentials.
	if err := creds.Save(credsPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving credentials: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Credentials saved to %s\n", credsPath)

	// Verify.
	fmt.Println("Verifying credentials...")
	if err := verify(creds); err != nil {
		fmt.Fprintf(os.Stderr, "Verification failed: %v\n", err)
		os.Exit(1)
	}
}

// cmdGet handles the "get" subcommand, dispatching to the appropriate resource handler.
func cmdGet(credsPath, consumerKey string, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: lp-cli get <resource> [flags]\n\nResources:\n  bug    Query and display a bug\n")
		os.Exit(1)
	}

	switch args[0] {
	case "bug":
		cmdGetBug(credsPath, consumerKey, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown resource: %s\n\nAvailable resources:\n  bug\n", args[0])
		os.Exit(1)
	}
}

// cmdGetBug handles "get bug -id <int>".
func cmdGetBug(credsPath, consumerKey string, args []string) {
	fs := flag.NewFlagSet("get bug", flag.ExitOnError)
	bugID := fs.Int("id", 0, "Bug ID (required)")
	fs.Parse(args)

	if *bugID <= 0 {
		fmt.Fprintf(os.Stderr, "Error: -id is required\n")
		fs.Usage()
		os.Exit(1)
	}

	creds, err := launchpad.LoadCredentials(credsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading credentials from %s: %v\n", credsPath, err)
		os.Exit(1)
	}
	client := launchpad.NewClient(creds, nil)

	if err := showBug(client, *bugID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// cmdSet handles the "set" subcommand, dispatching to the appropriate resource handler.
func cmdSet(credsPath, consumerKey string, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: lp-cli set <resource> [flags]\n\nResources:\n  bug    Update bug fields\n")
		os.Exit(1)
	}

	switch args[0] {
	case "bug":
		cmdSetBug(credsPath, consumerKey, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown resource: %s\n\nAvailable resources:\n  bug\n", args[0])
		os.Exit(1)
	}
}

// cmdSetBug handles "set bug -id <int> [-title <string>]".
func cmdSetBug(credsPath, consumerKey string, args []string) {
	fs := flag.NewFlagSet("set bug", flag.ExitOnError)
	bugID := fs.Int("id", 0, "Bug ID (required)")
	title := fs.String("title", "", "New bug title")
	// Future fields: add more flags here (e.g. -description, -importance).
	fs.Parse(args)

	if *bugID <= 0 {
		fmt.Fprintf(os.Stderr, "Error: -id is required\n")
		fs.Usage()
		os.Exit(1)
	}

	// Require at least one field to update.
	if *title == "" {
		fmt.Fprintf(os.Stderr, "Error: at least one field flag is required (e.g. -title)\n")
		fs.Usage()
		os.Exit(1)
	}

	creds, err := launchpad.LoadCredentials(credsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading credentials from %s: %v\n", credsPath, err)
		os.Exit(1)
	}
	client := launchpad.NewClient(creds, nil)

	if *title != "" {
		if err := updateBugTitle(client, *bugID, *title); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Display updated bug.
	if err := showBug(client, *bugID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// verify checks that the given credentials are valid by fetching the
// authenticated user's profile.
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

// showBug fetches and displays a bug by ID, including its tasks and assignees.
func showBug(client *launchpad.Client, bugID int) error {
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

	return nil
}
