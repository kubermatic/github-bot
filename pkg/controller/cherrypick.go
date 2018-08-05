package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/adtac/cherry-pick-bot/pkg/git"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

func (c *Controller) syncCherryPicks(ctx context.Context, repo github.Repository, id int) error {
	pr, _, err := c.client.PullRequests.Get(ctx, repo.GetOwner().GetLogin(), repo.GetName(), id)
	if err != nil {
		return fmt.Errorf("failed to fetch pr: %v", err)
	}
	if !pr.GetMerged() {
		glog.V(6).Infoln("PR not merged, nothing to do...")
		return nil
	}
	labels, _, err := c.client.Issues.ListLabelsByIssue(ctx, repo.GetOwner().GetLogin(), repo.GetName(), id, nil)
	if err != nil {
		fmt.Errorf("failed to fetch labels for issue %s: %v", repo.GetURL(), err)
	}

	for _, label := range labels {
		glog.V(6).Infof("Checking if label %s requires action", label.GetName())
		if strings.HasPrefix(label.GetName(), cherryPickLabelPrefix) {
			glog.Infof("Creating cherry-pick for label %s", label.String())
			if err := c.createCherryPick(ctx, repo, *pr, label.GetName()); err != nil {
				return fmt.Errorf("failed to create cherry pick: %v", err)
			}
		}
	}

	glog.V(6).Infoln("Successfully finished processing cherry-picks")
	return nil
}

func (c *Controller) createCherryPick(ctx context.Context, repo github.Repository, pr github.PullRequest, label string) error {
	splittedLabel := strings.Split(label, "/")
	if len(splittedLabel) < 2 {
		return fmt.Errorf("label %s is not of format 'cherry-pick/<branchname>'!", label)
	}
	baseBranch := strings.Join(splittedLabel[1:], "/")
	headBanch, err := git.PushCherryPick(repo.GetSSHURL(), baseBranch, pr.GetMergeCommitSHA())
	if err != nil {
		errWriteMessage := c.writeMessageToIssue(ctx, repo, pr.GetNumber(), fmt.Sprintf("Error creating cherry-pick due to: `%v`", err))
		if errWriteMessage != nil {
			return fmt.Errorf("error creating cherry-pick: %v, also writing an error message to github failed: %v",
				err, errWriteMessage)
		}
		return fmt.Errorf("error creating cherry-pick: %v", err)
	}

	title := fmt.Sprintf("Automated cherry-pick of %s onto %s", pr.GetTitle(), baseBranch)
	body := fmt.Sprintf("Automated cherry-pick of %s\n\n%s", pr.GetTitle(), pr.GetBody())
	pullRequest := &github.NewPullRequest{
		Base:  &baseBranch,
		Head:  &headBanch,
		Title: &title,
		Body:  &body,
	}

	newPR, _, err := c.client.PullRequests.Create(ctx, repo.GetOwner().GetLogin(), repo.GetName(), pullRequest)
	if err != nil {
		errWriteMessage := c.writeMessageToIssue(ctx, repo, pr.GetNumber(), fmt.Sprintf("Error creating pull request: %v", err))
		if errWriteMessage != nil {
			return fmt.Errorf("error creating pull request: %v, also writing an error message to github failed: %v", err, errWriteMessage)
		}
		return fmt.Errorf("error creating pull request: %v", err)
	}
	if err = c.writeMessageToIssue(ctx, repo, pr.GetNumber(), fmt.Sprintf("Created #%v to cherry-pick this pr onto %s", newPR.GetNumber(), baseBranch)); err != nil {
		return fmt.Errorf("failed to create success message after successfully creating PR: %v", err)
	}

	return nil
}
