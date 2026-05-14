package main

import (
	"fmt"

	"github.com/gkoh/launchpad"
	"github.com/spf13/cobra"
)

var (
	consumerKey string
	credsPath   string
)

var rootCmd = &cobra.Command{
	Use:   "lp-cli",
	Short: "CLI tool for interacting with the Launchpad API",
	Long: `lp-cli is a command-line tool for querying, searching, and updating
bugs on Launchpad. It uses OAuth 1.0a for authentication.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if credsPath == "" {
			p, err := launchpad.DefaultCredentialsPath(consumerKey)
			if err != nil {
				return err
			}
			credsPath = p
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&consumerKey, "consumer", "lp-cli", "OAuth consumer key (application name)")
	rootCmd.PersistentFlags().StringVar(&credsPath, "credentials", "", "Path to credentials file (default: ~/.config/launchpad/<consumer>.json)")
}

// newClient loads credentials and returns a configured Launchpad client.
func newClient() (*launchpad.Client, error) {
	creds, err := launchpad.LoadCredentials(credsPath)
	if err != nil {
		return nil, fmt.Errorf("loading credentials from %s: %w", credsPath, err)
	}
	return launchpad.NewClient(creds, nil), nil
}
