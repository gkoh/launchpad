# launchpad

A Go library and tools for interacting with the [Launchpad](https://launchpad.net) API.

## Overview

- **Library** (`github.com/gkoh/launchpad`) — stdlib-only Go package providing OAuth 1.0a authentication, typed API structs (bugs, bug tasks, branches, people, messages), and an HTTP client for the Launchpad REST API.
- **CLI** (`cmd/lp-cli`) — command-line tool for querying, searching, and updating bugs.
- **TUI** (`cmd/lp-tui`) — terminal UI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for interactive bug browsing.

## Install

```sh
go install github.com/gkoh/launchpad/cmd/lp-cli@latest
go install github.com/gkoh/launchpad/cmd/lp-tui@latest
```

## Quick start

Authenticate (opens a browser for OAuth authorization):

```sh
lp-cli auth
```

For write access:

```sh
lp-cli auth -permission WRITE_PRIVATE
```

Verify stored credentials:

```sh
lp-cli auth -check
```

### Querying bugs

```sh
lp-cli get bug -id 12345
lp-cli get bug -id 12345 -verbose   # include all comments
```

### Searching bugs

```sh
lp-cli search bug -project ubuntu -status "In Progress"
lp-cli search bug -project ubuntu -assignee username -importance High
lp-cli search bug -project ubuntu -tags "sru,kernel" -limit 20
lp-cli search bug -project ubuntu "search text"
```

### Updating bugs

All updates require `-project` to gate operations to a matching bug task.

```sh
# Update title
lp-cli set bug -id 12345 -project ubuntu -title "New title"

# Set assignee on a task
lp-cli set bug -id 12345 -project ubuntu -assignee username -task "ubuntu (Ubuntu)"

# Unassign
lp-cli set bug -id 12345 -project ubuntu -assignee "" -task "ubuntu (Ubuntu)"

# Set status
lp-cli set bug -id 12345 -project ubuntu -status "Fix Committed" -task "ubuntu (Ubuntu)"

# Set importance
lp-cli set bug -id 12345 -project ubuntu -importance High -task "ubuntu (Ubuntu)"
```

### TUI

```sh
lp-tui
```

Provides interactive search, list, and detail views for bugs. Requires credentials from `lp-cli auth`.

## Library usage

```go
package main

import (
	"fmt"
	"github.com/gkoh/launchpad"
)

func main() {
	creds, _ := launchpad.LoadCredentials("~/.config/lp-cli/credentials.json")
	client := launchpad.NewClient(creds, nil)

	resp, _ := client.Get("/bugs/1")
	defer resp.Body.Close()
	// ...
}
```

The library provides typed structs with `Link` fields (type-safe URL wrapper) for all Launchpad resource links, and typed string constants for enums like `BugTaskStatus`, `BugTaskImportance`, and `InformationType`.

## License

[MIT](LICENSE)
