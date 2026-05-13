package launchpad

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialsSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-creds.json")

	creds := &Credentials{
		ConsumerKey: "my-app",
		Token: &AccessToken{
			Token:  "tok123",
			Secret: "sec456",
		},
	}

	if err := creds.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file permissions.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file perm = %o, want 0600", perm)
	}

	loaded, err := LoadCredentials(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.ConsumerKey != creds.ConsumerKey {
		t.Errorf("ConsumerKey = %q, want %q", loaded.ConsumerKey, creds.ConsumerKey)
	}
	if loaded.Token.Token != creds.Token.Token {
		t.Errorf("Token = %q, want %q", loaded.Token.Token, creds.Token.Token)
	}
	if loaded.Token.Secret != creds.Token.Secret {
		t.Errorf("Secret = %q, want %q", loaded.Token.Secret, creds.Token.Secret)
	}
}

func TestLoadCredentialsNotFound(t *testing.T) {
	_, err := LoadCredentials("/nonexistent/path/creds.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCredentialsSaveCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "creds.json")

	creds := &Credentials{
		ConsumerKey: "app",
		Token:       &AccessToken{Token: "t", Secret: "s"},
	}
	if err := creds.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}
