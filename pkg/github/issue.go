package github

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

// IssueState defines the state of an issue
type IssueState string

const (
	//IssueStateOpen indicates the state of an open issue
	IssueStateOpen IssueState = "open"
	// IssueStateClosed indicates the state of a closed issue
	IssueStateClosed IssueState = "closed"
	// IssueStateAll indicates the state of all closed and open issues
	IssueStateAll IssueState = "all"
)

// IssueLister has the ability to list GitHub issues
type IssueLister struct {
	Organisation string
	Repository   string
	Token        string
}

// Issue keeps all the information returned for a GitHub issue
type Issue struct {
	ID           int64     `json:"id"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	State        string    `json:"state"`
	Repository   string    `json:"repository"`
	Organisation string    `json:"organisation"`
	Labels       []string  `json:"labels"`
	Body         string    `json:"body"`
	Comments     int       `json:"comments"`
	UserLogin    string    `json:"user_login"`
	Assignee     string    `json:"assignee"`
	ClosedBy     string    `json:"closed_by"`
	CreatedAt    time.Time `json:"created_at"`
	ClosedAt     time.Time `json:"closed_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	URL          string    `json:"url"`
}

// ToReader marshals the issue to JSON and returns a reader to resulted string
func (i *Issue) ToReader() (io.Reader, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, errors.Wrapf(err, "marshaling to JSON: %v", *i)
	}
	return strings.NewReader(string(data)), nil
}

// parseLabels parse the labels from a GitHub issue
func (i *IssueLister) parseLabels(issue *github.Issue) []string {
	result := []string{}
	for _, label := range issue.Labels {
		result = append(result, label.GetName())
	}
	return result
}

// toIssue converts a GitHub issue to Issue
func (i *IssueLister) toIssue(issue *github.Issue) *Issue {
	return &Issue{
		ID:           issue.GetID(),
		Number:       issue.GetNumber(),
		Title:        issue.GetTitle(),
		URL:          issue.GetURL(),
		State:        issue.GetState(),
		Repository:   i.Repository,
		Organisation: i.Organisation,
		Labels:       i.parseLabels(issue),
		Body:         issue.GetBody(),
		Comments:     issue.GetComments(),
		UserLogin:    issue.User.GetLogin(),
		Assignee:     issue.Assignee.GetLogin(),
		CreatedAt:    issue.GetCreatedAt(),
		ClosedAt:     issue.GetClosedAt(),
		UpdatedAt:    issue.GetUpdatedAt(),
		ClosedBy:     issue.ClosedBy.GetLogin(),
	}
}

// List list all open issues in a repository
func (i *IssueLister) List(ctx context.Context, state IssueState) ([]*Issue, error) {
	gh := client(ctx, i.Token)
	result := []*Issue{}
	opt := &github.IssueListByRepoOptions{
		State: string(state),
	}
	issues, _, err := gh.Issues.ListByRepo(ctx, i.Organisation, i.Repository, opt)
	if err != nil {
		return result, errors.Wrapf(err, "listing all issues in %s/%s", i.Organisation, i.Repository)
	}
	for _, issue := range issues {
		if !issue.IsPullRequest() {
			result = append(result, i.toIssue(issue))
		}
	}
	return result, nil
}

// ListStream list all GitHub issues using pagination and streaming them in a batch into the
// channel. If any error occurs, it is returned into the error channel.
func (i *IssueLister) ListStream(ctx context.Context, pageSize int, state IssueState) (chan []*Issue, chan error) {
	errCh := make(chan error)
	issuesCh := make(chan []*Issue)
	go func() {
		gh := client(ctx, i.Token)
		opt := &github.IssueListByRepoOptions{
			ListOptions: github.ListOptions{
				PerPage: pageSize,
			},
			State: string(state),
		}
		for {
			issues, resp, err := gh.Issues.ListByRepo(ctx, i.Organisation, i.Repository, opt)
			if err != nil {
				errCh <- err
				break
			}
			result := []*Issue{}
			for _, issue := range issues {
				if !issue.IsPullRequest() {
					result = append(result, i.toIssue(issue))
				}
			}
			select {
			case <-ctx.Done():
				break
			case issuesCh <- result:
			}
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
		close(errCh)
		close(issuesCh)
	}()
	return issuesCh, errCh
}
