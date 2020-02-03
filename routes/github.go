// Copyright 2019 Christian Banse
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"issues"
	"math"
	"net/http"
	"strings"

	"github.com/google/go-github/v29/github"
)

func (router *Router) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	eventType := r.Header.Get("X-Github-Event")

	var (
		clients *issues.GitHubClients
		err     error
	)

	if eventType == "issue_comment" {
		var (
			event github.IssueCommentEvent
		)

		decoder.Decode(&event)

		if clients, err = router.app.GetInstallationClients(event.GetInstallation().GetID()); err != nil {
			log.Errorf("Could not create installation client: %s", err)
			return
		}

		log.Debugf("Got event %s for issue comment in %s", event.GetAction(), issues.GetIssueIdentifier(event.GetRepo(), event.GetIssue()))

		if event.GetAction() == "created" {
			comment := event.Comment.GetBody()

			if strings.HasPrefix(comment, "/branch") {
				router.handleBranchIssue(clients, event)
				return
			}
		}
	} else if eventType == "issues" {
		var (
			event github.IssuesEvent
		)
		decoder.Decode(&event)

		if clients, err = router.app.GetInstallationClients(event.GetInstallation().GetID()); err != nil {
			log.Errorf("Could not create installation client: %s", err)
			return
		}

		log.Debugf("Got event %s for issue %s", event.GetAction(), issues.GetIssueIdentifier(event.Repo, event.Issue))

		if event.GetAction() == "edited" {
			// do not trigger on bot updates, otherwise we will update forever
			if event.Sender.GetType() == "Bot" {
				return
			}

			router.handleIssueChange(clients, event)
		}
	} else {
		log.Warnf("Not handling unknown event type %s", eventType)
	}
}

func shortIssueTitle(issue *github.Issue) (shortTitle string) {
	if issue == nil {
		return ""
	}

	// lowercase
	shortTitle = strings.ToLower(issue.GetTitle())

	// split words
	words := strings.Split(shortTitle, " ")

	// join back together with '-' (4 max)
	shortTitle = strings.Join(words[0:int(math.Min(5, float64(len(words))))], "-")

	return
}

func (router *Router) handleBranchIssue(clients *issues.GitHubClients, event github.IssueCommentEvent) {
	var (
		err        error
		resp       *github.Response
		issue      *github.Issue
		repo       *github.Repository
		ref        *github.Reference
		branch     *github.Branch
		branchName string
	)

	// desired branch name
	issue = event.GetIssue()
	repo = event.GetRepo()
	branchName = fmt.Sprintf("%d-%s", issue.GetNumber(), shortIssueTitle(issue))

	log.Infof("Issue %s is now developed in branch %s", issues.GetIssueIdentifier(repo, issue), branchName)

	if branch, resp, err = clients.V3.Repositories.GetBranch(context.Background(), repo.GetOwner().GetLogin(), repo.GetName(), branchName); err != nil {
		if resp == nil || resp != nil && resp.StatusCode != 404 {
			log.Errorf("Retrieving branch %s from %s failed: %s", branchName, issues.GetIssueIdentifier(repo, issue), err)
			return
		}
	}

	if branch != nil {
		log.Debugf("Branch %s already exists", branchName)
		return
	}

	baseRef := fmt.Sprintf("heads/%s", repo.GetDefaultBranch())

	// need to get the current ref from default branch
	if ref, _, err = clients.V3.Git.GetRef(context.Background(), repo.GetOwner().GetLogin(), repo.GetName(), baseRef); err != nil {
		log.Errorf("Retrieving ref %s failed: %s", baseRef, err)
		return
	}

	refString := fmt.Sprintf("refs/heads/%s", branchName)
	ref.Ref = &refString

	// and push a new ref with the new branch name
	if ref, _, err = clients.V3.Git.CreateRef(context.Background(), repo.GetOwner().GetLogin(), repo.GetName(), ref); err != nil {
		log.Errorf("Creating ref for branch %s failed: %s", branchName, err)
		return
	}

	log.Debugf("Created branch %s (%s) for issue %s", branchName, ref.GetRef(), issues.GetIssueIdentifier(repo, issue))

	body := fmt.Sprintf("Created branch [%s](/%s/%s/tree/%s) for development of this issue", branchName, repo.GetOwner().GetLogin(), repo.GetName(), branchName)

	if _, _, err = clients.V3.Issues.CreateComment(context.Background(), repo.GetOwner().GetLogin(), repo.GetName(), issue.GetNumber(), &github.IssueComment{
		Body: &body,
	}); err != nil {
		log.Errorf("Creating comment for issue %s failed: %s", issues.GetIssueIdentifier(repo, issue), err)
		return
	}

	/*body := fmt.Sprintf("Fixes #%d", issue.GetNumber())
	issueNumber := issue.GetNumber()
	base := repo.GetDefaultBranch()
	modify := true
	draft := true

	pull := &github.NewPullRequest{
		Head:                &branchName,
		Base:                &base,
		Body:                &body,
		Issue:               &issueNumber,
		MaintainerCanModify: &modify,
		Draft:               &draft,
	}

	if _, _, err = clients.V3.PullRequests.Create(context.Background(), repo.GetOwner().GetLogin(), repo.GetName(), pull); err != nil {
		log.Errorf("Creating the pull request for %s failed: %s", issues.GetIssueIdentifier(repo, issue), err)
		return
	}*/
}

func (router *Router) handleIssueChange(clients *issues.GitHubClients, event github.IssuesEvent) {
	var err error

	issue := event.GetIssue()

	// update epic status, if necessary
	router.app.UpdateEpicStatus(clients, event)

	// find relationships to other issues
	var relationships []issues.Relationship
	if _, err = router.app.GetDatabase().Select(&relationships, "select * from relationship where \"issueId\"=$1 or \"otherIssueId\"=$2", issue.GetNumber(), issue.GetNumber()); err != nil {
		log.Errorf("Could not fetch relationships to other issues from database: %s", err)
		return
	}

	// skip, if there are no relationships
	if len(relationships) == 0 {
		log.Debugf("Issue %s does not have relationships, not updating", issues.GetIssueIdentifier(event.GetRepo(), event.GetIssue()))
		return
	}

	footer := "\n\n---\n\n"

	for _, relationship := range relationships {
		footer += fmt.Sprintf("**Issue %s #%d**\n", relationship.Type, relationship.OtherIssueID)
	}

	// TODO: find out if its already there and just edit the footer
	newBody := fmt.Sprintf("%s%s", *issue.Body, footer)
	request := github.IssueRequest{
		Body: &newBody,
	}

	// update issue text
	if _, _, err = clients.V3.Issues.Edit(context.Background(), *event.Repo.Owner.Login, *event.Repo.Name, *issue.Number, &request); err != nil {
		log.Errorf("Updating issue %s failed: %s", issues.GetIssueIdentifier(event.GetRepo(), event.GetIssue()), err)
		return
	}

	log.Infof("Updated issue %s", issues.GetIssueIdentifier(event.GetRepo(), event.GetIssue()))
}
