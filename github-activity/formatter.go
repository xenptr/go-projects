package main

import (
	"fmt"
	"strings"
)

func formatEvent(e Event) string {
	switch e.Type {
	case CommitCommentEvent:
		return fmt.Sprintf("- Commented on a commit in %s\n", e.Repo.Name)

	case CreateEvent:
		if e.Payload.RefType == "repository" {
			return fmt.Sprintf("- Created repository %s\n", e.Repo.Name)
		}
		return fmt.Sprintf("- Created %s %s in %s\n", e.Payload.RefType, e.Payload.Ref, e.Repo.Name)

	case DeleteEvent:
		return fmt.Sprintf("- Deleted %s %s from %s\n", e.Payload.RefType, e.Payload.Ref, e.Repo.Name)

	case ForkEvent:
		return fmt.Sprintf("- Forked %s\n", e.Repo.Name)

	case IssueCommentEvent:
		return fmt.Sprintf("- Commented on issue #%d in %s\n", e.Payload.Issue.Number, e.Repo.Name)

	case IssuesEvent:
		switch e.Payload.Action {
		case "opened", "closed", "reopened":
			return fmt.Sprintf("- %s issue #%d in %s\n", capitalize(e.Payload.Action), e.Payload.Issue.Number, e.Repo.Name)
		case "assigned":
			return fmt.Sprintf("- %s %s to issue #%d in %s\n", capitalize(e.Payload.Action), e.Payload.Assignee.Login, e.Payload.Issue.Number, e.Repo.Name)
		case "unassigned":
			return fmt.Sprintf("- %s %s from issue #%d in %s\n", capitalize(e.Payload.Action), e.Payload.Assignee.Login, e.Payload.Issue.Number, e.Repo.Name)
		case "labeled":
			return fmt.Sprintf("- Added label %q to issue #%d in %s\n", e.Payload.Label.Name, e.Payload.Issue.Number, e.Repo.Name)
		case "unlabeled":
			return fmt.Sprintf("- Removed label %q from issue #%d in %s\n", e.Payload.Label.Name, e.Payload.Issue.Number, e.Repo.Name)
		default:
			return ""
		}

	case MemberEvent:
		switch e.Payload.Action {
		case "added":
			return fmt.Sprintf("- %s %s as a collaborator to %s\n", capitalize(e.Payload.Action), e.Payload.Member.Login, e.Repo.Name)
		case "removed":
			return fmt.Sprintf("- %s %s as a collaborator from %s\n", capitalize(e.Payload.Action), e.Payload.Member.Login, e.Repo.Name)
		default:
			return fmt.Sprintf("- %s %s in %s\n", capitalize(e.Payload.Action), e.Payload.Member.Login, e.Repo.Name)
		}

	case PullRequestEvent:
		switch e.Payload.Action {
		case "opened", "closed", "reopened":
			if e.Payload.Action == "closed" && e.Payload.PullRequest.Merged {
				return fmt.Sprintf("- Merged PR #%d in %s\n", e.Payload.Number, e.Repo.Name)
			}
			return fmt.Sprintf("- %s PR #%d in %s\n", capitalize(e.Payload.Action), e.Payload.Number, e.Repo.Name)
		case "assigned":
			return fmt.Sprintf("- %s %s to PR #%d in %s\n", capitalize(e.Payload.Action), e.Payload.Assignee.Login, e.Payload.Number, e.Repo.Name)
		case "unassigned":
			return fmt.Sprintf("- %s %s from PR #%d in %s\n", capitalize(e.Payload.Action), e.Payload.Assignee.Login, e.Payload.Number, e.Repo.Name)
		case "labeled":
			return fmt.Sprintf("- Added label %q to PR #%d in %s\n", e.Payload.Label.Name, e.Payload.Number, e.Repo.Name)
		case "unlabeled":
			return fmt.Sprintf("- Removed label %q from PR #%d in %s\n", e.Payload.Label.Name, e.Payload.Number, e.Repo.Name)
		default:
			return ""
		}

	case PullRequestReviewEvent:
		switch e.Payload.Action {
		case "created":
			switch e.Payload.Review.State {
			case "approved":
				return fmt.Sprintf("- Approved PR #%d in %s\n", e.Payload.PullRequest.Number, e.Repo.Name)
			case "changes_requested":
				return fmt.Sprintf("- Requested changes on PR #%d in %s\n", e.Payload.PullRequest.Number, e.Repo.Name)
			default:
				return fmt.Sprintf("- Reviewed PR #%d in %s\n", e.Payload.PullRequest.Number, e.Repo.Name)
			}
		case "updated", "dismissed":
			return fmt.Sprintf("- %s review on PR #%d in %s\n", capitalize(e.Payload.Action), e.Payload.PullRequest.Number, e.Repo.Name)
		default:
			return ""
		}

	case PullRequestReviewCommentEvent:
		return fmt.Sprintf("- Added review comment to PR #%d in %s\n", e.Payload.PullRequest.Number, e.Repo.Name)

	case PushEvent:
		return fmt.Sprintf("- Pushed changes to %s\n", e.Repo.Name)

	case ReleaseEvent:
		return fmt.Sprintf("- %s release %s in %s\n", capitalize(e.Payload.Action), e.Payload.Release.TagName, e.Repo.Name)

	case WatchEvent:
		return fmt.Sprintf("- Starred %s\n", e.Repo.Name)

	default:
		return fmt.Sprintf("- %s occurred in %s\n", e.Type, e.Repo.Name)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func formatUser(u User) string {
	name := u.Name
	if name == "" {
		name = u.Login
	}

	var b strings.Builder

	fmt.Fprintln(&b, strings.Repeat("─", 50))
	fmt.Fprintf(&b, "  %s (%s)\n", name, u.Login)

	if u.Bio != "" {
		fmt.Fprintf(&b, "  %s\n", u.Bio)
	}
	if u.Company != "" {
		fmt.Fprintf(&b, "  Company   : %s\n", u.Company)
	}
	if u.Location != "" {
		fmt.Fprintf(&b, "  Location  : %s\n", u.Location)
	}

	fmt.Fprintf(&b, "  Repos     : %d\n", u.PublicRepos)
	fmt.Fprintf(&b, "  Followers : %d  Following: %d\n", u.Followers, u.Following)
	fmt.Fprintf(&b, "  %s\n", u.HTMLURL)
	fmt.Fprintln(&b, strings.Repeat("─", 50))

	return b.String()
}
