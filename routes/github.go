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
	"issues/db"
	"net/http"

	"github.com/google/go-github/github"
)

func GitHubCallback(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	eventType := r.Header.Get("X-Github-Event")

	if eventType == "issues" {
		var (
			event   github.IssuesEvent
			clients *issues.GitHubClients
			err     error
		)
		decoder.Decode(&event)

		if clients, err = issues.GetInstallationClients(event.GetInstallation().GetID()); err != nil {
			return
		}

		if event.GetAction() == "edited" {
			// do not trigger on bot updates, otherwise we will update forever
			if event.Sender.GetType() == "Bot" {
				return
			}

			handleIssueChange(clients, event)
		}
	} else {
		log.Warnf("Not handling unknown event type %s", eventType)
	}

}

func handleIssueChange(clients *issues.GitHubClients, event github.IssuesEvent) {
	var err error

	issue := event.GetIssue()

	// find relationships to other issues
	var relationships []issues.Relationship
	if _, err = db.Select(&relationships, "select * from relationship where \"issueId\"=$1 or \"otherIssueId\"=$2", issue.GetNumber(), issue.GetNumber()); err != nil {
		log.Errorf("Could not fetch relationships to other issues from database: %s", err)
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
		log.Errorf("Updating issue %s/%s#%d failed: %s", *event.Repo.Owner.Login, *event.Repo.Name, *issue.Number, err)
		return
	}

	log.Infof("Updated issue %s/%s#%d.", *event.Repo.Owner.Login, *event.Repo.Name, *issue.Number)
}
