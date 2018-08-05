package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/adtac/cherry-pick-bot/pkg/git"

	"github.com/google/go-github/github"
)

func (c *Controller) syncCherryPicks(ctx context.Context, repo github.Repository, id int) error {
	pr, _, err := c.client.PullRequests.Get(ctx, repo.GetOwner().GetLogin(), repo.GetName(), id)
	if err != nil {
		return fmt.Errorf("failed to fetch pr: %v", err)
	}
	if !pr.GetMerged() {
		return nil
	}
	labels, _, err := c.client.Issues.ListLabelsByIssue(ctx, repo.GetOwner().GetLogin(), repo.GetName(), id, nil)
	if err != nil {
		fmt.Errorf("failed to fetch labels for issue %s: %v", repo.GetURL(), err)
	}

	for _, label := range labels {
		if strings.HasPrefix(label.String(), cherryPickLabelPrefix) {
			if err := c.createCherryPick(ctx, repo, *pr, label.String()); err != nil {
				return fmt.Errorf("failed to create cherry pick: %v", err)
			}
		}
	}

	return nil
}

func (c *Controller) createCherryPick(ctx context.Context, repo github.Repository, pr github.PullRequest, label string) error {
	splittedLabel := strings.Split(label, "/")
	if len(splittedLabel) != 2 {
		return fmt.Errorf("label %s is not of format 'cherry-pick/<branchname>'!", label)
	}
	headBanch, err := git.PushCherryPick(repo.GetURL(), splittedLabel[1], pr.GetMergeCommitSHA())
	if err != nil {
		errWriteMessage := c.writeMessageToIssue(ctx, repo, int(pr.GetID()), fmt.Sprintf("Error creating cherry-pick: %v", err))
		if errWriteMessage != nil {
			return fmt.Errorf("error creating cherry-pick: %v, also writing an error message to github failed: %v",
				err, errWriteMessage)
		}
		return fmt.Errorf("error creating cherry-pick: %v", err)
	}

	title := fmt.Sprintf("Automated cherry-pick of %s onto %s", pr.GetTitle(), splittedLabel[1])
	body := fmt.Sprintf("Automated cherry-pick of %s\n----\n%s", pr.GetTitle(), pr.GetBody())
	pullRequest := &github.NewPullRequest{
		Base:  &splittedLabel[1],
		Head:  &headBanch,
		Title: &title,
		Body:  &body,
	}

	_, _, err = c.client.PullRequests.Create(ctx, repo.GetOwner().GetLogin(), repo.GetName(), pullRequest)
	if err != nil {
		errWriteMessage := c.writeMessageToIssue(ctx, repo, int(pr.GetID()), fmt.Sprintf("Error creating pull request: %v", err))
		if errWriteMessage != nil {
			return fmt.Errorf("error creating pull request: %v, also writing an error message to github failed: %v", err, errWriteMessage)
		}
		return fmt.Errorf("error creating pull request: %v", err)
	}

	return nil
}

//Create(ctx context.Context, owner string, repo string, pull *NewPullRequest) (*PullRequest, *Response, error)
