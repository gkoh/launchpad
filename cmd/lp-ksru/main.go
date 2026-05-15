// Command lp-ksru searches for kernel SRU workflow bugs by SRU cycle tag.
package main

import (
	"fmt"
	"os"
	"regexp"
	"time"

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
	showSnaps   bool
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
	rootCmd.Flags().BoolVar(&showSnaps, "snaps", false, "Include snap bugs (bugs with snap-prepare task)")
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

	// Deduplicate by bug ID — multiple tasks may reference the same bug.
	seen := make(map[int]bool)
	var uniqueBugIDs []int
	for _, task := range tasks {
		bugID, err := task.BugID()
		if err != nil {
			continue
		}
		if seen[bugID] {
			continue
		}
		seen[bugID] = true
		uniqueBugIDs = append(uniqueBugIDs, bugID)
	}

	fmt.Printf("# SRU Cycle: %s\n", sruCycle)

	// Fetch each bug and display its details.
	displayed := 0
outer:
	for _, bugID := range uniqueBugIDs {
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

		// Skip snap bugs unless --snaps is set.
		if !showSnaps {
			isSnap := false
			for _, t := range bugTasks {
				if t.BugTargetName == sruProject+"/snap-prepare" {
					isSnap = true
					break
				}
			}
			if isSnap {
				continue outer
			}
		}

		// Skip bugs where the kernel-sru-workflow task is Fix Committed.
		for _, t := range bugTasks {
			if t.BugTargetName == sruProject && t.Status == launchpad.BugTaskStatusFixCommitted {
				continue outer
			}
		}

		displayed++
		fmt.Printf("\n## Bug #%d: %s\n\n", bug.ID, bug.Title)
		fmt.Printf("- **Tasks:**\n")
		for _, t := range bugTasks {
			if t.Status == launchpad.BugTaskStatusIncomplete {
				if t.DateIncomplete != nil {
					fmt.Printf("  - %s: %s (%s)\n", t.BugTargetName, t.Status, humanDuration(time.Since(*t.DateIncomplete)))
				} else {
					fmt.Printf("  - %s: %s\n", t.BugTargetName, t.Status)
				}
			}
		}
		fmt.Printf("- **Web:** %s\n", bug.WebLink)
	}

	fmt.Printf("\nDisplayed %d of %d bugs\n", displayed, len(uniqueBugIDs))

	return nil
}

// searchAllTasks fetches all bug tasks matching the SRU cycle tag,
// following pagination links.
func searchAllTasks(client *launchpad.Client, sruCycle string) ([]launchpad.BugTask, error) {
	var tags []string
	for i := 1; i <= maxSRUSuffix; i++ {
		tags = append(tags, fmt.Sprintf("kernel-sru-cycle-%s-%d", sruCycle, i))
	}

	return client.SearchTasks(sruProject, &launchpad.SearchTasksOptions{
		Tags:           tags,
		TagsCombinator: launchpad.TagsCombinatorAny,
		PageSize:       pageSize,
		FollowPages:    true,
	})
}

// humanDuration formats a duration into a human-readable string using
// the two largest meaningful units.
func humanDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	const (
		day   = 24 * time.Hour
		week  = 7 * day
		month = 30 * day
		year  = 365 * day
	)

	switch {
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm", m)
	case d < day:
		h := int(d.Hours())
		return fmt.Sprintf("%dh", h)
	case d < week:
		days := int(d / day)
		return fmt.Sprintf("%dd", days)
	case d < month:
		weeks := int(d / week)
		days := int((d % week) / day)
		if days > 0 {
			return fmt.Sprintf("%dw %dd", weeks, days)
		}
		return fmt.Sprintf("%dw", weeks)
	case d < year:
		months := int(d / month)
		weeks := int((d % month) / week)
		if weeks > 0 {
			return fmt.Sprintf("%dmo %dw", months, weeks)
		}
		return fmt.Sprintf("%dmo", months)
	default:
		years := int(d / year)
		months := int((d % year) / month)
		if months > 0 {
			return fmt.Sprintf("%dy %dmo", years, months)
		}
		return fmt.Sprintf("%dy", years)
	}
}
