// Command lp-ksru searches for kernel SRU workflow bugs by SRU cycle tag.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gkoh/launchpad"
	"github.com/spf13/cobra"
)

const (
	// sruProject is the Launchpad project for kernel SRU workflow tracking.
	sruProject = "kernel-sru-workflow"

	// pageSize is the number of results per API page.
	pageSize = 75
)

// maxSRUSuffix is the maximum -N suffix appended to SRU cycle tags.
const maxSRUSuffix = 20

// sruCyclePattern validates the YYYY.MM.DD or sYYYY.MM.DD format.
var sruCyclePattern = regexp.MustCompile(`^s?\d{4}\.\d{2}\.\d{2}$`)

var (
	consumerKey string
	credsPath   string
)

var rootCmd = &cobra.Command{
	Use:   "lp-ksru <sru-cycle>",
	Short: "Search kernel SRU workflow bugs by SRU cycle tag",
	Long: `lp-ksru searches the kernel-sru-workflow project on Launchpad for all
bugs tagged with the given SRU cycle (format: YYYY.MM.DD).`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          run,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&consumerKey, "consumer", "lp-cli", "OAuth consumer key (application name)")
	rootCmd.PersistentFlags().StringVar(&credsPath, "credentials", "", "Path to credentials file (default: ~/.config/launchpad/<consumer>.json)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newClient() (*launchpad.Client, error) {
	if credsPath == "" {
		p, err := launchpad.DefaultCredentialsPath(consumerKey)
		if err != nil {
			return nil, err
		}
		credsPath = p
	}

	creds, err := launchpad.LoadCredentials(credsPath)
	if err != nil {
		return nil, fmt.Errorf("loading credentials from %s: %w", credsPath, err)
	}
	return launchpad.NewClient(creds, nil), nil
}

func run(cmd *cobra.Command, args []string) error {
	sruCycle := args[0]
	if !sruCyclePattern.MatchString(sruCycle) {
		return fmt.Errorf("invalid sru-cycle %q; expected format YYYY.MM.DD or sYYYY.MM.DD (e.g. 2025.05.12 or s2025.05.12)", sruCycle)
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	// Fetch all matching bug tasks across all pages.
	tasks, err := searchAllTasks(client, sruCycle)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Printf("No bugs found for SRU cycle %s\n", sruCycle)
		return nil
	}

	// Deduplicate by bug link — multiple tasks may reference the same bug.
	type bugEntry struct {
		bugLink string
		order   int // preserve first-seen order
	}
	seen := make(map[string]bool)
	var uniqueBugLinks []string
	for _, task := range tasks {
		link := task.BugLink.String()
		if link == "" || seen[link] {
			continue
		}
		seen[link] = true
		uniqueBugLinks = append(uniqueBugLinks, link)
	}

	fmt.Printf("# SRU Cycle: %s (%d bugs)\n", sruCycle, len(uniqueBugLinks))

	// Fetch each bug and display its details.
	for _, bugLink := range uniqueBugLinks {
		bugID, err := parseBugID(bugLink)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not parse bug ID from %s: %v\n", bugLink, err)
			continue
		}

		bug, err := client.GetBug(bugID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not fetch bug #%d: %v\n", bugID, err)
			continue
		}

		bugTasks, err := bug.GetTasks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not fetch tasks for bug #%d: %v\n", bugID, err)
			continue
		}

		fmt.Printf("\n## Bug #%d: %s\n\n", bug.ID, bug.Title)
		if len(bug.Tags) > 0 {
			fmt.Printf("- **Tags:** %s\n", strings.Join(bug.Tags, ", "))
		}
		if len(bugTasks) > 0 {
			fmt.Printf("- **Tasks:**\n")
			for _, t := range bugTasks {
				fmt.Printf("  - %s: %s\n", t.BugTargetName, t.Status)
			}
		}
		fmt.Printf("- **Web:** %s\n", bug.WebLink)
	}

	return nil
}

// searchAllTasks fetches all bug tasks matching the SRU cycle tag,
// following pagination links.
func searchAllTasks(client *launchpad.Client, sruCycle string) ([]launchpad.BugTask, error) {
	params := url.Values{}
	params.Set("ws.op", "searchTasks")
	for i := 1; i <= maxSRUSuffix; i++ {
		params.Add("tags", fmt.Sprintf("kernel-sru-cycle-%s-%d", sruCycle, i))
	}
	params.Set("tags_combinator", "Any")
	params.Set("ws.size", fmt.Sprintf("%d", pageSize))

	searchURL := fmt.Sprintf("/%s?%s", sruProject, params.Encode())

	var all []launchpad.BugTask

	resp, err := client.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("project %q not found", sruProject)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var collection launchpad.BugTaskCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	all = append(all, collection.Entries...)

	// Follow pagination.
	nextURL := collection.NextCollectionLink.String()
	for nextURL != "" {
		resp, err := client.GetAbsolute(nextURL)
		if err != nil {
			return all, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return all, fmt.Errorf("reading response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return all, fmt.Errorf("API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var page launchpad.BugTaskCollection
		if err := json.Unmarshal(body, &page); err != nil {
			return all, fmt.Errorf("parsing response: %w", err)
		}

		all = append(all, page.Entries...)
		nextURL = page.NextCollectionLink.String()
	}

	return all, nil
}

// parseBugID extracts the numeric bug ID from a Launchpad bug API link
// (e.g. "https://api.launchpad.net/devel/bugs/12345" → 12345).
func parseBugID(bugLink string) (int, error) {
	idx := strings.LastIndex(bugLink, "/")
	if idx < 0 || idx == len(bugLink)-1 {
		return 0, fmt.Errorf("no path segment in %q", bugLink)
	}
	return strconv.Atoi(bugLink[idx+1:])
}
