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
//	  bug -id <int> -project <name> [-title <s>] [-assignee <user> -task <target>] [-status <s> -task <target>] [-importance <s> -task <target>]
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gkoh/launchpad"
)

// statusContentReturned is Launchpad's non-standard HTTP 209 response
// indicating a successful PATCH with content returned.
const statusContentReturned = 209

// maxDescriptionLength is the maximum number of characters to display
// for a bug description before truncating.
const maxDescriptionLength = 200

// maxCommentLength is the maximum number of characters to display
// for a comment before truncating.
const maxCommentLength = 500

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
	case "search":
		cmdSearch(*credsPath, *consumerKey, args[1:])
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
    bug -id <int> [-verbose]     Query and display a bug by ID

  set <resource> [flags]                          Update a resource
    bug -id <int> -project <name> [field flags]   Update bug fields
      -title <string>                             New bug title
      -assignee <user>                            Set assignee (use "" to unassign)
      -status <status>                            Set bug task status
      -task <target>                              Bug task target (required with -assignee, -status)

  search bug [flags] [<text>]      Search for bugs
    -project <name>              Project or distribution (required)
    -status <status>             Filter by status
    -importance <importance>     Filter by importance
    -assignee <name>             Filter by assignee username
    -tags <t1,t2,...>            Filter by tags (comma-separated)
    -limit <int>                 Max results (default: 10)
    <text>                       Full-text search filter (positional)
`)
}

// cmdAuth handles the "auth" subcommand.
func cmdAuth(credsPath, consumerKey string, args []string) {
	fs := flag.NewFlagSet("auth", flag.ExitOnError)
	check := fs.Bool("check", false, "Verify stored credentials and exit")
	permission := fs.String("permission", launchpad.PermissionReadPrivate, "Permission level (READ_PUBLIC, READ_PRIVATE, WRITE_PUBLIC, WRITE_PRIVATE)")
	fs.Parse(args)

	// Detect whether -permission was explicitly provided.
	permissionSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "permission" {
			permissionSet = true
		}
	})

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
	// Skip reuse when -permission is explicitly set, since the user
	// wants a new token with different permissions.
	if !permissionSet {
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

// cmdGetBug handles "get bug -id <int> [-verbose]".
func cmdGetBug(credsPath, consumerKey string, args []string) {
	fs := flag.NewFlagSet("get bug", flag.ExitOnError)
	bugID := fs.Int("id", 0, "Bug ID (required)")
	verbose := fs.Bool("verbose", false, "Show all comments")
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

	if err := showBug(client, *bugID, *verbose); err != nil {
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

// cmdSetBug handles "set bug -id <int> -project <name> [-title <string>] [-assignee <user> -task <target>] [-status <s> -task <target>]".
func cmdSetBug(credsPath, consumerKey string, args []string) {
	fs := flag.NewFlagSet("set bug", flag.ExitOnError)
	bugID := fs.Int("id", 0, "Bug ID (required)")
	project := fs.String("project", "", "Project or distribution name (required)")
	title := fs.String("title", "", "New bug title")
	assignee := fs.String("assignee", "", "Assignee username (use \"\" to unassign)")
	status := fs.String("status", "", "Bug task status (e.g. New, Confirmed, \"In Progress\", \"Fix Committed\")")
	importance := fs.String("importance", "", "Bug task importance (e.g. Critical, High, Medium)")
	task := fs.String("task", "", "Bug task target name (required with -assignee, -status, -importance)")
	fs.Parse(args)

	if *bugID <= 0 {
		fmt.Fprintf(os.Stderr, "Error: -id is required\n")
		fs.Usage()
		os.Exit(1)
	}

	if *project == "" {
		fmt.Fprintf(os.Stderr, "Error: -project is required\n")
		fs.Usage()
		os.Exit(1)
	}

	// Detect which flags were explicitly set.
	assigneeSet := false
	statusSet := false
	importanceSet := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "assignee":
			assigneeSet = true
		case "status":
			statusSet = true
		case "importance":
			importanceSet = true
		}
	})

	// Require at least one field to update.
	if *title == "" && !assigneeSet && !statusSet && !importanceSet {
		fmt.Fprintf(os.Stderr, "Error: at least one field flag is required (e.g. -title, -assignee, -status, -importance)\n")
		fs.Usage()
		os.Exit(1)
	}

	// -assignee, -status, and -importance require -task.
	if (assigneeSet || statusSet || importanceSet) && *task == "" {
		fmt.Fprintf(os.Stderr, "Error: -task is required when using -assignee, -status, or -importance\n")
		fs.Usage()
		os.Exit(1)
	}

	// Validate -status value locally.
	if statusSet {
		if !validBugTaskStatus(*status) {
			fmt.Fprintf(os.Stderr, "Error: invalid status %q; valid values: %s\n", *status, validBugTaskStatusList())
			os.Exit(1)
		}
	}

	// Validate -importance value locally.
	if importanceSet {
		if !validBugTaskImportance(*importance) {
			fmt.Fprintf(os.Stderr, "Error: invalid importance %q; valid values: %s\n", *importance, validBugTaskImportanceList())
			os.Exit(1)
		}
	}

	creds, err := launchpad.LoadCredentials(credsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading credentials from %s: %v\n", credsPath, err)
		os.Exit(1)
	}
	client := launchpad.NewClient(creds, nil)

	// Gate: fetch bug and tasks, verify the bug has a task matching -project.
	resp, err := client.Get(fmt.Sprintf("/bugs/%d", *bugID))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading bug: %v\n", err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: API returned %d: %s\n", resp.StatusCode, strings.TrimSpace(string(body)))
		os.Exit(1)
	}

	var bug launchpad.Bug
	if err := json.Unmarshal(body, &bug); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing bug: %v\n", err)
		os.Exit(1)
	}

	if bug.BugTasksCollectionLink.IsZero() {
		fmt.Fprintf(os.Stderr, "Error: bug #%d has no tasks\n", *bugID)
		os.Exit(1)
	}

	tasks, err := fetchBugTasks(client, bug.BugTasksCollectionLink.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching tasks: %v\n", err)
		os.Exit(1)
	}

	if _, err := findTaskByTarget(tasks, *project); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *title != "" {
		if err := updateBugTitle(client, *bugID, *title); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	if assigneeSet {
		if err := updateBugTaskAssignee(client, *bugID, tasks, *task, *assignee); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	if statusSet {
		if err := updateBugTaskStatus(client, *bugID, tasks, *task, *status); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	if importanceSet {
		if err := updateBugTaskImportance(client, *bugID, tasks, *task, *importance); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Display updated bug.
	if err := showBug(client, *bugID, false); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// cmdSearch handles the "search" subcommand.
func cmdSearch(credsPath, consumerKey string, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: lp-cli search <resource> [flags]\n\nResources:\n  bug    Search for bugs\n")
		os.Exit(1)
	}

	switch args[0] {
	case "bug":
		cmdSearchBug(credsPath, consumerKey, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown resource: %s\n\nAvailable resources:\n  bug\n", args[0])
		os.Exit(1)
	}
}

// cmdSearchBug handles "search bug [flags]".
func cmdSearchBug(credsPath, consumerKey string, args []string) {
	fs := flag.NewFlagSet("search bug", flag.ExitOnError)
	project := fs.String("project", "", "Project or distribution name (required)")
	status := fs.String("status", "", "Filter by bug task status (e.g. New, Confirmed, In Progress)")
	importance := fs.String("importance", "", "Filter by importance (e.g. Critical, High, Medium)")
	assignee := fs.String("assignee", "", "Filter by assignee username")
	tags := fs.String("tags", "", "Filter by tags (comma-separated)")
	limit := fs.Int("limit", 10, "Maximum number of results")
	fs.Parse(args)

	if *project == "" {
		fmt.Fprintf(os.Stderr, "Error: -project is required\n")
		fs.Usage()
		os.Exit(1)
	}

	// Remaining positional arguments form the search text.
	searchText := strings.Join(fs.Args(), " ")

	// Build the search URL with ws.op=searchTasks.
	params := url.Values{}
	params.Set("ws.op", "searchTasks")
	if searchText != "" {
		params.Set("search_text", searchText)
	}
	if *status != "" {
		params.Set("status", *status)
	}
	if *importance != "" {
		params.Set("importance", *importance)
	}
	if *assignee != "" {
		params.Set("assignee", fmt.Sprintf("https://api.launchpad.net/devel/~%s", *assignee))
	}
	if *tags != "" {
		for _, tag := range strings.Split(*tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				params.Add("tags", tag)
			}
		}
	}
	params.Set("ws.size", fmt.Sprintf("%d", *limit))

	path := fmt.Sprintf("/%s?%s", *project, params.Encode())

	creds, err := launchpad.LoadCredentials(credsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading credentials from %s: %v\n", credsPath, err)
		os.Exit(1)
	}
	client := launchpad.NewClient(creds, nil)

	resp, err := client.Get(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode == http.StatusNotFound {
		fmt.Fprintf(os.Stderr, "Error: project %q not found\n", *project)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "API returned %d: %s\n", resp.StatusCode, strings.TrimSpace(string(body)))
		os.Exit(1)
	}

	var collection launchpad.BugTaskCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if len(collection.Entries) == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Printf("# Search Results (%d)\n", len(collection.Entries))
	for _, task := range collection.Entries {
		fmt.Printf("\n## %s\n\n", task.Title)
		fmt.Printf("- **Target:** %s\n", task.BugTargetDisplayName)
		fmt.Printf("- **Status:** %s\n", task.Status)
		fmt.Printf("- **Importance:** %s\n", task.Importance)
		if !task.AssigneeLink.IsZero() {
			fmt.Printf("- **Assignee:** %s\n", task.AssigneeLink)
		}
		fmt.Printf("- **Web:** %s\n", task.WebLink)
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var user struct {
		DisplayName string         `json:"display_name"`
		Name        string         `json:"name"`
		WebLink     launchpad.Link `json:"web_link"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fmt.Printf("Authenticated as: %s (%s)\n", user.DisplayName, user.Name)
	fmt.Printf("Profile: %s\n", user.WebLink)
	return nil
}

// showBug fetches and displays a bug by ID, including its tasks and assignees.
// When verbose is true, all comments are fetched and displayed.
func showBug(client *launchpad.Client, bugID int, verbose bool) error {
	resp, err := client.Get(fmt.Sprintf("/bugs/%d", bugID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var bug launchpad.Bug
	if err := json.Unmarshal(body, &bug); err != nil {
		return fmt.Errorf("parsing bug: %w", err)
	}

	// Display bug summary.
	fmt.Printf("# Bug #%d: %s\n\n", bug.ID, bug.Title)
	fmt.Printf("- **Info type:** %s\n", bug.InformationType)
	fmt.Printf("- **Lock status:** %s\n", bug.LockStatus)
	fmt.Printf("- **Heat:** %d\n", bug.Heat)
	if len(bug.Tags) > 0 {
		fmt.Printf("- **Tags:** %s\n", strings.Join(bug.Tags, ", "))
	}
	if bug.DateCreated != nil {
		fmt.Printf("- **Created:** %s\n", bug.DateCreated.Format("2006-01-02 15:04:05"))
	}
	if bug.DateLastUpdated != nil {
		fmt.Printf("- **Updated:** %s\n", bug.DateLastUpdated.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("- **Messages:** %d\n", bug.MessageCount)
	fmt.Printf("- **Web:** %s\n", bug.WebLink)

	if bug.Description != "" {
		desc := bug.Description
		if len(desc) > maxDescriptionLength {
			desc = desc[:maxDescriptionLength] + "..."
		}
		fmt.Printf("\n## Description\n\n%s\n", desc)
	}

	// Fetch bug tasks.
	if !bug.BugTasksCollectionLink.IsZero() {
		tasks, err := fetchBugTasks(client, bug.BugTasksCollectionLink.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nWarning: could not fetch bug tasks: %v\n", err)
		} else if len(tasks) > 0 {
			// Resolve assignees.
			assignees := resolveAssignees(client, tasks)

			fmt.Printf("\n## Tasks\n")
			for _, task := range tasks {
				fmt.Printf("\n### %s\n\n", task.BugTargetDisplayName)
				fmt.Printf("- **Status:** %s\n", task.Status)
				fmt.Printf("- **Importance:** %s\n", task.Importance)
				if name, ok := assignees[task.AssigneeLink.String()]; ok {
					fmt.Printf("- **Assignee:** %s\n", name)
				} else {
					fmt.Printf("- **Assignee:** unassigned\n")
				}
			}
		}
	}

	// Fetch and display comments when verbose.
	if verbose && !bug.MessagesCollectionLink.IsZero() {
		messages, err := fetchAllMessages(client, bug.MessagesCollectionLink.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nWarning: could not fetch comments: %v\n", err)
		} else if len(messages) > 0 {
			// Resolve owner display names.
			owners := resolveOwners(client, messages)

			fmt.Printf("\n## Comments (%d)\n", len(messages))
			for i, msg := range messages {
				owner := msg.OwnerLink.String()
				if name, ok := owners[msg.OwnerLink.String()]; ok {
					owner = name
				}
				date := "unknown"
				if msg.DateCreated != nil {
					date = msg.DateCreated.Format("2006-01-02 15:04:05")
				}
				fmt.Printf("\n### #%d by %s on %s\n\n", i+1, owner, date)
				content := msg.Content
				if len(content) > maxCommentLength {
					content = content[:maxCommentLength] + "..."
				}
				if content != "" {
					fmt.Printf("%s\n", content)
				}
			}
		}
	}

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

	if resp.StatusCode != http.StatusOK {
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
		if task.AssigneeLink.IsZero() {
			continue
		}
		link := task.AssigneeLink.String()
		if _, ok := result[link]; ok {
			continue // already resolved
		}

		req, err := http.NewRequest(http.MethodGet, link, nil)
		if err != nil {
			result[link] = link
			continue
		}
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			result[link] = link
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusOK {
			result[link] = link
			continue
		}

		var person launchpad.Person
		if err := json.Unmarshal(body, &person); err != nil {
			result[link] = link
			continue
		}

		result[link] = fmt.Sprintf("%s (%s)", person.DisplayName, person.Name)
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

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("not authorized to edit bug #%d (credentials may lack write permission; re-run 'lp-cli auth -permission WRITE_PRIVATE')", bugID)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != statusContentReturned {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

// updateBugTaskAssignee sets the assignee on a bug task matching the given
// target name. An empty username unassigns the task. The tasks slice must
// have been fetched beforehand.
func updateBugTaskAssignee(client *launchpad.Client, bugID int, tasks []launchpad.BugTask, targetName, username string) error {
	matched, err := findTaskByTarget(tasks, targetName)
	if err != nil {
		return err
	}

	// Build the PATCH payload.
	var payload []byte
	if username == "" {
		// Unassign: send null.
		payload = []byte(`{"assignee_link":null}`)
	} else {
		assigneeLink := fmt.Sprintf("https://api.launchpad.net/devel/~%s", username)
		payload, err = json.Marshal(map[string]string{"assignee_link": assigneeLink})
		if err != nil {
			return fmt.Errorf("marshalling payload: %w", err)
		}
	}

	// PATCH the task.
	taskURL := matched.SelfLink.String()
	req, err := http.NewRequest(http.MethodPatch, taskURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	patchResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("updating assignee: %w", err)
	}
	defer patchResp.Body.Close()

	if patchResp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("not authorized to edit bug #%d (credentials may lack write permission; re-run 'lp-cli auth -permission WRITE_PRIVATE')", bugID)
	}
	if patchResp.StatusCode != http.StatusOK && patchResp.StatusCode != statusContentReturned {
		patchBody, _ := io.ReadAll(patchResp.Body)
		return fmt.Errorf("API returned %d: %s", patchResp.StatusCode, strings.TrimSpace(string(patchBody)))
	}

	if username == "" {
		fmt.Printf("Unassigned task %q on bug #%d\n", targetName, bugID)
	} else {
		fmt.Printf("Assigned %q to task %q on bug #%d\n", username, targetName, bugID)
	}
	return nil
}

// findTaskByTarget returns the bug task whose BugTargetName matches the given
// name exactly. If no match is found, it returns an error listing available
// target names.
func findTaskByTarget(tasks []launchpad.BugTask, targetName string) (*launchpad.BugTask, error) {
	for i := range tasks {
		if tasks[i].BugTargetName == targetName {
			return &tasks[i], nil
		}
	}
	var names []string
	for _, t := range tasks {
		names = append(names, t.BugTargetName)
	}
	return nil, fmt.Errorf("no task found for target %q; available targets: %s", targetName, strings.Join(names, ", "))
}

// updateBugTaskStatus sets the status on a bug task matching the given target
// name. The tasks slice must have been fetched beforehand.
func updateBugTaskStatus(client *launchpad.Client, bugID int, tasks []launchpad.BugTask, targetName, status string) error {
	matched, err := findTaskByTarget(tasks, targetName)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]string{"status": status})
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}

	taskURL := matched.SelfLink.String()
	req, err := http.NewRequest(http.MethodPatch, taskURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("updating status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("not authorized to edit bug #%d (credentials may lack write permission; re-run 'lp-cli auth -permission WRITE_PRIVATE')", bugID)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != statusContentReturned {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	fmt.Printf("Set status %q on task %q for bug #%d\n", status, targetName, bugID)
	return nil
}

// validBugTaskStatus returns true if s is a valid BugTaskStatus value.
func validBugTaskStatus(s string) bool {
	switch launchpad.BugTaskStatus(s) {
	case launchpad.BugTaskStatusNew,
		launchpad.BugTaskStatusIncomplete,
		launchpad.BugTaskStatusOpinion,
		launchpad.BugTaskStatusInvalid,
		launchpad.BugTaskStatusWontFix,
		launchpad.BugTaskStatusExpired,
		launchpad.BugTaskStatusConfirmed,
		launchpad.BugTaskStatusTriaged,
		launchpad.BugTaskStatusInProgress,
		launchpad.BugTaskStatusDeferred,
		launchpad.BugTaskStatusFixCommitted,
		launchpad.BugTaskStatusFixReleased,
		launchpad.BugTaskStatusDoesNotExist,
		launchpad.BugTaskStatusUnknown:
		return true
	}
	return false
}

// validBugTaskStatusList returns a comma-separated string of all valid
// BugTaskStatus values for use in error messages.
func validBugTaskStatusList() string {
	return strings.Join([]string{
		string(launchpad.BugTaskStatusNew),
		string(launchpad.BugTaskStatusIncomplete),
		string(launchpad.BugTaskStatusOpinion),
		string(launchpad.BugTaskStatusInvalid),
		string(launchpad.BugTaskStatusWontFix),
		string(launchpad.BugTaskStatusExpired),
		string(launchpad.BugTaskStatusConfirmed),
		string(launchpad.BugTaskStatusTriaged),
		string(launchpad.BugTaskStatusInProgress),
		string(launchpad.BugTaskStatusDeferred),
		string(launchpad.BugTaskStatusFixCommitted),
		string(launchpad.BugTaskStatusFixReleased),
		string(launchpad.BugTaskStatusDoesNotExist),
		string(launchpad.BugTaskStatusUnknown),
	}, ", ")
}

// updateBugTaskImportance sets the importance on a bug task matching the given
// target name. The tasks slice must have been fetched beforehand.
func updateBugTaskImportance(client *launchpad.Client, bugID int, tasks []launchpad.BugTask, targetName, importance string) error {
	matched, err := findTaskByTarget(tasks, targetName)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]string{"importance": importance})
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}

	taskURL := matched.SelfLink.String()
	req, err := http.NewRequest(http.MethodPatch, taskURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("updating importance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("not authorized to edit bug #%d (credentials may lack write permission; re-run 'lp-cli auth -permission WRITE_PRIVATE')", bugID)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != statusContentReturned {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	fmt.Printf("Set importance %q on task %q for bug #%d\n", importance, targetName, bugID)
	return nil
}

// validBugTaskImportance returns true if s is a valid BugTaskImportance value.
func validBugTaskImportance(s string) bool {
	switch launchpad.BugTaskImportance(s) {
	case launchpad.BugTaskImportanceUnknown,
		launchpad.BugTaskImportanceUndecided,
		launchpad.BugTaskImportanceCritical,
		launchpad.BugTaskImportanceHigh,
		launchpad.BugTaskImportanceMedium,
		launchpad.BugTaskImportanceLow,
		launchpad.BugTaskImportanceWishlist:
		return true
	}
	return false
}

// validBugTaskImportanceList returns a comma-separated string of all valid
// BugTaskImportance values for use in error messages.
func validBugTaskImportanceList() string {
	return strings.Join([]string{
		string(launchpad.BugTaskImportanceUnknown),
		string(launchpad.BugTaskImportanceUndecided),
		string(launchpad.BugTaskImportanceCritical),
		string(launchpad.BugTaskImportanceHigh),
		string(launchpad.BugTaskImportanceMedium),
		string(launchpad.BugTaskImportanceLow),
		string(launchpad.BugTaskImportanceWishlist),
	}, ", ")
}

// fetchAllMessages fetches all messages for a bug, following pagination links.
func fetchAllMessages(client *launchpad.Client, url string) ([]launchpad.Message, error) {
	var all []launchpad.Message

	for url != "" {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return all, err
		}
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return all, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return all, err
		}

		if resp.StatusCode != http.StatusOK {
			return all, fmt.Errorf("messages returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var collection launchpad.MessageCollection
		if err := json.Unmarshal(body, &collection); err != nil {
			return all, err
		}

		all = append(all, collection.Entries...)
		url = collection.NextCollectionLink.String()
	}

	return all, nil
}

// resolveOwners fetches the display name for each unique owner link in messages.
// Returns a map from owner_link URL to "DisplayName (name)" string.
func resolveOwners(client *launchpad.Client, messages []launchpad.Message) map[string]string {
	result := make(map[string]string)

	for _, msg := range messages {
		if msg.OwnerLink.IsZero() {
			continue
		}
		link := msg.OwnerLink.String()
		if _, ok := result[link]; ok {
			continue
		}

		req, err := http.NewRequest(http.MethodGet, link, nil)
		if err != nil {
			result[link] = link
			continue
		}
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			result[link] = link
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusOK {
			result[link] = link
			continue
		}

		var person launchpad.Person
		if err := json.Unmarshal(body, &person); err != nil {
			result[link] = link
			continue
		}

		result[link] = fmt.Sprintf("%s (%s)", person.DisplayName, person.Name)
	}

	return result
}
