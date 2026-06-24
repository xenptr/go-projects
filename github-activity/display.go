package main

import (
	"fmt"
	"strings"
)

func formatEvent(event Event) string {
	switch event.Type {
	case CommitCommentEvent:
		return fmt.Sprintf("- Commented on a commit in %s\n", event.Repo.Name)

	case CreateEvent:
		if event.Payload.RefType == "repository" {
			return fmt.Sprintf("- Created repository %s\n", event.Repo.Name)
		}
		return fmt.Sprintf("- Created %s %s in %s\n", event.Payload.RefType, event.Payload.Ref, event.Repo.Name)

	case DeleteEvent:
		return fmt.Sprintf("- Deleted %s %s from %s\n", event.Payload.RefType, event.Payload.Ref, event.Repo.Name)

	case ForkEvent:
		return fmt.Sprintf("- Forked %s\n", event.Repo.Name)

	case IssueCommentEvent:
		return fmt.Sprintf("- Commented on issue #%d in %s\n", event.Payload.Issue.Number, event.Repo.Name)

	case IssuesEvent:
		switch event.Payload.Action {
		case "opened", "closed", "reopened":
			return fmt.Sprintf("- %s issue #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Issue.Number, event.Repo.Name)
		case "assigned":
			return fmt.Sprintf("- %s %s to issue #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Assignee.Login, event.Payload.Issue.Number, event.Repo.Name)
		case "unassigned":
			return fmt.Sprintf("- %s %s from issue #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Assignee.Login, event.Payload.Issue.Number, event.Repo.Name)
		case "labeled":
			return fmt.Sprintf("- Added label %q to issue #%d in %s\n", event.Payload.Label.Name, event.Payload.Issue.Number, event.Repo.Name)
		case "unlabeled":
			return fmt.Sprintf("- Removed label %q from issue #%d in %s\n", event.Payload.Label.Name, event.Payload.Issue.Number, event.Repo.Name)
		default:
			return ""
		}

	case MemberEvent:
		switch event.Payload.Action {
		case "added":
			return fmt.Sprintf("- %s %s as a collaborator to %s\n", capitalize(event.Payload.Action), event.Payload.Member.Login, event.Repo.Name)
		case "removed":
			return fmt.Sprintf("- %s %s as a collaborator from %s\n", capitalize(event.Payload.Action), event.Payload.Member.Login, event.Repo.Name)
		default:
			return fmt.Sprintf("- %s %s in %s\n", capitalize(event.Payload.Action), event.Payload.Member.Login, event.Repo.Name)
		}

	case PullRequestEvent:
		switch event.Payload.Action {
		case "opened", "closed", "reopened":
			if event.Payload.Action == "closed" && event.Payload.PullRequest.Merged {
				return fmt.Sprintf("- Merged PR #%d in %s\n", event.Payload.Number, event.Repo.Name)
			}
			return fmt.Sprintf("- %s PR #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Number, event.Repo.Name)
		case "assigned":
			return fmt.Sprintf("- %s %s to PR #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Assignee.Login, event.Payload.Number, event.Repo.Name)
		case "unassigned":
			return fmt.Sprintf("- %s %s from PR #%d in %s\n", capitalize(event.Payload.Action), event.Payload.Assignee.Login, event.Payload.Number, event.Repo.Name)
		case "labeled":
			return fmt.Sprintf("- Added label %q to PR #%d in %s\n", event.Payload.Label.Name, event.Payload.Number, event.Repo.Name)
		case "unlabeled":
			return fmt.Sprintf("- Removed label %q from PR #%d in %s\n", event.Payload.Label.Name, event.Payload.Number, event.Repo.Name)
		default:
			return ""
		}

	case PullRequestReviewEvent:
		switch event.Payload.Action {
		case "created":
			switch event.Payload.Review.State {
			case "approved":
				return fmt.Sprintf("- Approved PR #%d in %s\n", event.Payload.PullRequest.Number, event.Repo.Name)
			case "changes_requested":
				return fmt.Sprintf("- Requested changes on PR #%d in %s\n", event.Payload.PullRequest.Number, event.Repo.Name)
			default:
				return fmt.Sprintf("- Reviewed PR #%d in %s\n", event.Payload.PullRequest.Number, event.Repo.Name)
			}
		case "updated", "dismissed":
			return fmt.Sprintf("- %s review on PR #%d in %s\n", capitalize(event.Payload.Action), event.Payload.PullRequest.Number, event.Repo.Name)
		default:
			return ""
		}

	case PullRequestReviewCommentEvent:
		return fmt.Sprintf("- Added review comment to PR #%d in %s\n", event.Payload.PullRequest.Number, event.Repo.Name)

	case PushEvent:
		return fmt.Sprintf("- Pushed changes to %s\n", event.Repo.Name)

	case ReleaseEvent:
		return fmt.Sprintf("- %s release %s in %s\n", capitalize(event.Payload.Action), event.Payload.Release.TagName, event.Repo.Name)

	case WatchEvent:
		return fmt.Sprintf("- Starred %s\n", event.Repo.Name)

	default:
		return fmt.Sprintf("- %s occurred in %s\n", event.Type, event.Repo.Name)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}
