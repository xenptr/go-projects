package main

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

var EventTypes = []EventType{
	CommitCommentEvent, CreateEvent, DeleteEvent, ForkEvent,
	IssueCommentEvent, IssuesEvent, MemberEvent, PullRequestEvent,
	PullRequestReviewEvent, PullRequestReviewCommentEvent,
	PushEvent, ReleaseEvent, WatchEvent,
}

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
