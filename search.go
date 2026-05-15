package launchpad

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TagsCombinator controls how multiple tags are matched in a search.
type TagsCombinator string

const (
	TagsCombinatorAll TagsCombinator = "All"
	TagsCombinatorAny TagsCombinator = "Any"
)

// SearchTasksOptions configures a bug task search.
type SearchTasksOptions struct {
	SearchText     string
	Status         string
	Importance     string
	Assignee       string // Launchpad username; converted to full API link internally.
	Tags           []string
	TagsCombinator TagsCombinator // Defaults to TagsCombinatorAll if empty.
	PageSize       int            // ws.size; 0 uses the API default.
	FollowPages    bool           // When true, all pages are fetched automatically.
}

// SearchTasks searches for bug tasks in the given project.
// If opts is nil, all tasks for the project are returned (subject to API defaults).
func (c *Client) SearchTasks(project string, opts *SearchTasksOptions) ([]BugTask, error) {
	params := url.Values{}
	params.Set("ws.op", "searchTasks")

	if opts != nil {
		if opts.SearchText != "" {
			params.Set("search_text", opts.SearchText)
		}
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Importance != "" {
			params.Set("importance", opts.Importance)
		}
		if opts.Assignee != "" {
			params.Set("assignee", fmt.Sprintf("%s/~%s", c.apiBase(), opts.Assignee))
		}
		if len(opts.Tags) > 0 {
			for _, tag := range opts.Tags {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					params.Add("tags", tag)
				}
			}
		}
		if opts.TagsCombinator != "" {
			params.Set("tags_combinator", string(opts.TagsCombinator))
		}
		if opts.PageSize > 0 {
			params.Set("ws.size", fmt.Sprintf("%d", opts.PageSize))
		}
	}

	path := fmt.Sprintf("/%s?%s", project, params.Encode())

	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("launchpad: reading search response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("launchpad: project %q not found", project)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("launchpad: search returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var collection BugTaskCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("launchpad: parsing search response: %w", err)
	}

	all := collection.Entries

	if opts != nil && opts.FollowPages {
		nextURL := collection.NextCollectionLink.String()
		for nextURL != "" {
			resp, err := c.GetAbsolute(nextURL)
			if err != nil {
				return all, err
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return all, fmt.Errorf("launchpad: reading search page: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				return all, fmt.Errorf("launchpad: search page returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			}

			var page BugTaskCollection
			if err := json.Unmarshal(body, &page); err != nil {
				return all, fmt.Errorf("launchpad: parsing search page: %w", err)
			}

			all = append(all, page.Entries...)
			nextURL = page.NextCollectionLink.String()
		}
	}

	for i := range all {
		all[i].client = c
	}

	return all, nil
}
