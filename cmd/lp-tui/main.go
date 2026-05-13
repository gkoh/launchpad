// Command lp-tui is a terminal UI for browsing Launchpad bugs.
//
// Usage:
//
//	lp-tui [-consumer <key>] [-credentials <path>]
//
// Requires credentials from lp-cli auth.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gkoh/launchpad"
)

func main() {
	consumerKey := flag.String("consumer", "lp-cli", "OAuth consumer key (application name)")
	credsPath := flag.String("credentials", "", "Path to credentials file (default: ~/.config/launchpad/<consumer>.json)")
	flag.Parse()

	if *credsPath == "" {
		p, err := launchpad.DefaultCredentialsPath(*consumerKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		*credsPath = p
	}

	creds, err := launchpad.LoadCredentials(*credsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading credentials from %s: %v\n", *credsPath, err)
		fmt.Fprintf(os.Stderr, "Run 'lp-cli auth' first to authenticate.\n")
		os.Exit(1)
	}

	client := launchpad.NewClient(creds, nil)

	p := tea.NewProgram(
		newModel(client),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
