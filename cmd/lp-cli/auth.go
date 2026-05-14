package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gkoh/launchpad"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Launchpad",
	Long:  "Authenticate with Launchpad using OAuth 1.0a. Opens a browser for authorization.",
	RunE:  runAuth,
}

var (
	authCheck      bool
	authPermission string
)

func init() {
	authCmd.Flags().BoolVar(&authCheck, "check", false, "Verify stored credentials and exit")
	authCmd.Flags().StringVar(&authPermission, "permission", launchpad.PermissionReadPrivate, "Permission level (READ_PUBLIC, READ_PRIVATE, WRITE_PUBLIC, WRITE_PRIVATE)")
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, args []string) error {
	if authCheck {
		creds, err := launchpad.LoadCredentials(credsPath)
		if err != nil {
			return fmt.Errorf("loading credentials from %s: %w", credsPath, err)
		}
		return verify(creds)
	}

	// Check for existing valid credentials.
	// Skip reuse when -permission is explicitly set, since the user
	// wants a new token with different permissions.
	permissionSet := cmd.Flags().Changed("permission")
	if !permissionSet {
		if creds, err := launchpad.LoadCredentials(credsPath); err == nil {
			fmt.Printf("Found existing credentials at %s\n", credsPath)
			fmt.Println("Verifying...")
			if err := verify(creds); err != nil {
				fmt.Fprintf(os.Stderr, "Existing credentials are invalid: %v\n", err)
				fmt.Println("Re-authenticating...")
			} else {
				return nil
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
		return err
	}

	// Step 2: User authorizes.
	authURL := launchpad.AuthorizeURL(cfg, reqToken, authPermission)
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
		return err
	}

	creds := &launchpad.Credentials{
		ConsumerKey: consumerKey,
		Token:       accessToken,
	}

	// Save credentials.
	if err := creds.Save(credsPath); err != nil {
		return fmt.Errorf("saving credentials: %w", err)
	}
	fmt.Printf("Credentials saved to %s\n", credsPath)

	// Verify.
	fmt.Println("Verifying credentials...")
	return verify(creds)
}

// verify checks that the given credentials are valid by fetching the
// authenticated user's profile.
func verify(creds *launchpad.Credentials) error {
	client := launchpad.NewClient(creds, nil)
	person, err := client.Me()
	if err != nil {
		return err
	}

	fmt.Printf("Authenticated as: %s (%s)\n", person.DisplayName, person.Name)
	fmt.Printf("Profile: %s\n", person.WebLink)
	return nil
}
