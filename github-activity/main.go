package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"unicode"
)

type EventType string

const (
	CommitCommentEvent            EventType = "CommitCommentEvent"
	CreateEvent                   EventType = "CreateEvent"
	DeleteEvent                   EventType = "DeleteEvent"
	ForkEvent                     EventType = "ForkEvent"
	IssueCommentEvent             EventType = "IssueCommentEvent"
	IssuesEvent                   EventType = "IssuesEvent"
	MemberEvent                   EventType = "MemberEvent"
	PullRequestEvent              EventType = "PullRequestEvent"
	PullRequestReviewEvent        EventType = "PullRequestReviewEvent"
	PullRequestReviewCommentEvent EventType = "PullRequestReviewCommentEvent"
	PushEvent                     EventType = "PushEvent"
	ReleaseEvent                  EventType = "ReleaseEvent"
	WatchEvent                    EventType = "WatchEvent"
)

type Event struct {
	Type    EventType `json:"type"`
	Repo    Repo      `json:"repo"`
	Payload Payload   `json:"payload"`

	// Payload json.RawMessage `json:"payload"`
}

type Repo struct {
	Name string `json:"name"`
}

type Payload struct {
	Action string `json:"action"`
	Number int    `json:"number"`

	Ref     string `json:"ref"`
	RefType string `json:"ref_type"`

	Member      Member      `json:"member"`
	Issue       Issue       `json:"issue"`
	Assignee    Assignee    `json:"assignee"`
	Label       Label       `json:"label"`
	PullRequest PullRequest `json:"pull_request"`
	Review      Review      `json:"review"`
	Release     Release     `json:"release"`
}

type Member struct {
	Login string `json:"login"`
}

type Issue struct {
	Number int `json:"number"`
}

type Assignee struct {
	Login string `json:"login"`
}

type Label struct {
	Name string `json:"name"`
}

type PullRequest struct {
	Number int  `json:"number"`
	Merged bool `json:"merged"`
}

type Review struct {
	State string `json:"state"`
}

type Release struct {
	TagName string `json:"tag_name"`
}

func main() {
	var username string

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./github-activity <username>")
		os.Exit(1)
	}

	username = os.Args[1]

	githubApi := fmt.Sprintf("https://api.github.com/users/%s/events", username)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, githubApi, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Github API returned %s\n", resp.Status)
		os.Exit(1)
	}

	var events []Event

	err = json.NewDecoder(resp.Body).Decode(&events)
	if err != nil {
		panic(err)
	}

	if len(events) == 0 {
		fmt.Printf("No recent public activity found for %s\n", username)
		return
	}

	for _, event := range events {
		switch event.Type {
		case CommitCommentEvent:
			fmt.Printf("- Commented on a commit in %s\n", event.Repo.Name)

		case CreateEvent:
			if event.Payload.RefType == "repository" {
				fmt.Printf("- Created repository %s\n", event.Repo.Name)
				continue
			}
			fmt.Printf("- Created %s %s in %s\n", event.Payload.RefType, event.Payload.Ref, event.Repo.Name)

		case DeleteEvent:
			fmt.Printf("- Deleted %s %s from %s\n", event.Payload.RefType, event.Payload.Ref, event.Repo.Name)

		case ForkEvent:
			fmt.Printf("- Forked %s\n", event.Repo.Name)

		case IssueCommentEvent:
			fmt.Printf("- Commented on issue #%d in %s\n", event.Payload.Issue.Number, event.Repo.Name)

		case IssuesEvent:
			switch event.Payload.Action {
			case "opened", "closed", "reopened":
				fmt.Printf("- %s issue #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Issue.Number, event.Repo.Name)
			case "assigned":
				fmt.Printf("- %s %s to issue #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Assignee.Login, event.Payload.Issue.Number, event.Repo.Name)
			case "unassigned":
				fmt.Printf("- %s %s from issue #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Assignee.Login, event.Payload.Issue.Number, event.Repo.Name)
			case "labeled":
				fmt.Printf("- Added label %q to issue #%d in %s\n", event.Payload.Label.Name, event.Payload.Issue.Number, event.Repo.Name)
			case "unlabeled":
				fmt.Printf("- Removed label %q from issue #%d in %s\n", event.Payload.Label.Name, event.Payload.Issue.Number, event.Repo.Name)
			default:
				continue
			}

		case MemberEvent:
			switch event.Payload.Action {
			case "added":
				fmt.Printf("- %s %s as a collaborator to %s\n", capitalize(event.Payload.Action), event.Payload.Member.Login, event.Repo.Name)
			case "removed":
				fmt.Printf("- %s %s as a collaborator from %s\n", capitalize(event.Payload.Action), event.Payload.Member.Login, event.Repo.Name)
			default:
				fmt.Printf("- %s %s in %s\n", capitalize(event.Payload.Action), event.Payload.Member.Login, event.Repo.Name)
			}

		case PullRequestEvent:
			switch event.Payload.Action {
			case "opened", "closed", "reopened":
				if event.Payload.Action == "closed" && event.Payload.PullRequest.Merged {
					fmt.Printf("- Merged PR #%d in %s\n", event.Payload.Number, event.Repo.Name)
					continue
				}
				fmt.Printf("- %s PR #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Number, event.Repo.Name)
			case "assigned":
				fmt.Printf("- %s %s to PR #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Assignee.Login, event.Payload.Number, event.Repo.Name)
			case "unassigned":
				fmt.Printf("- %s %s from PR #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Assignee.Login, event.Payload.Number, event.Repo.Name)
			case "labeled":
				fmt.Printf("- Added label %q to PR #%d in %s\n", event.Payload.Label.Name, event.Payload.Number, event.Repo.Name)
			case "unlabeled":
				fmt.Printf("- Removed label %q from PR #%d in %s\n", event.Payload.Label.Name, event.Payload.Number, event.Repo.Name)
			default:
				continue
			}

		case PullRequestReviewEvent:
			switch event.Payload.Action {
			case "created":
				switch event.Payload.Review.State {
				case "approved":
					fmt.Printf("- Approved PR #%d in %s\n", event.Payload.PullRequest.Number, event.Repo.Name)
				case "changes_requested":
					fmt.Printf("- Requested changes on PR #%d in %s\n", event.Payload.PullRequest.Number, event.Repo.Name)
				default:
					fmt.Printf("- Reviewed PR #%d in %s\n", event.Payload.PullRequest.Number, event.Repo.Name)
				}
			case "updated", "dismissed":
				fmt.Printf("- %s review on PR #%d in %s\n", capitalize(event.Payload.Action), event.Payload.PullRequest.Number, event.Repo.Name)
			default:
				continue
			}

		case PullRequestReviewCommentEvent:
			fmt.Printf("- Added review comment to PR #%d in %s\n", event.Payload.PullRequest.Number, event.Repo.Name)

		case PushEvent:
			fmt.Printf("- Pushed changes to %s\n", event.Repo.Name)

		case ReleaseEvent:
			fmt.Printf("- %s release %s in %s\n", capitalize(event.Payload.Action), event.Payload.Release.TagName, event.Repo.Name)

		case WatchEvent:
			fmt.Printf("- Starred %s\n", event.Repo.Name)

		default:
			fmt.Printf("- %s occurred in %s\n", event.Type, event.Repo.Name)
		}
	}

	// enc := json.NewEncoder(os.Stdout)
	// enc.SetIndent("", "  ")
	// enc.Encode(events)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
