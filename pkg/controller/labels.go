package controller

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

// For the labels part we have to treat it as an issue, because PRs do not have Labels in this lib
func (c *Controller) syncLabels(ctx context.Context, repo github.Repository, id int, body string) error {
	labelsToEnsure := getLabelsToEnsureFromMessage(body)
	currentLabels, _, err := c.client.Issues.ListLabelsByIssue(ctx, repo.GetOwner().GetLogin(), repo.GetName(), id, nil)
	if err != nil {
		fmt.Errorf("failed to fetch labels for issue %s: %v", repo.GetURL(), err)
	}

	for _, desiredLabel := range labelsToEnsure {
		if !labelSliceContains(currentLabels, desiredLabel) {
			// AddLabelsToIssue(ctx context.Context, owner string, repo string, number int, labels []string) ([]*Label, *Response, error)
			newLabels, _, err := c.client.Issues.AddLabelsToIssue(ctx, repo.GetOwner().GetLogin(), repo.GetName(), id, []string{desiredLabel})
			if err != nil {
				return fmt.Errorf("failed to add label %s on issue %s#%v: %v", desiredLabel, repo.GetURL(), id, err)
			}
			glog.V(6).Infof("Successfully added label %s on issue %s#%v", desiredLabel, repo.GetURL(), id)
			currentLabels = newLabels
		}
	}
	return nil
}

func getLabelsToEnsureFromMessage(message string) []string {
	var labels []string
	if cherryPickCommandTarget := getCommandTarget(message, cherryPickCommandName); cherryPickCommandTarget != nil {
		labels = append(labels, fmt.Sprintf("%s/%s", cherryPickLabelPrefix, *cherryPickCommandTarget))
	}

	return labels
}

func labelSliceContains(slice []*github.Label, s string) bool {
	for _, element := range slice {
		if element.GetName() == s {
			return true
		}
	}
	return false
}
