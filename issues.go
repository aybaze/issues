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

package issues

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jsternberg/markdownfmt/markdown"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	oauth2GitHub "golang.org/x/oauth2/github"
	"gopkg.in/russross/blackfriday.v2"

	"github.com/google/go-github/github"
)

var (
	log *logrus.Entry
)

type Application struct {
	AppID int64
	db    Database
	gh    *oauth2.Config
}

func init() {
	log = logrus.WithField("component", "main")
}

func NewApplication(appID int64, db Database) *Application {
	app := Application{appID, db, nil}

	db.Init()

	return &app
}

// GetDatabase directly exposes the database to outside the application.
// TODO: remove this call and just expose the GetXYZ calls to directly get workspaces, issues, etc.
func (app *Application) GetDatabase() Database {
	return app.db
}

// AddServiceConnection adds a new connection to an external service. For now this
// just statically sets up the GitHub connection
func (app *Application) AddServiceConnection(service string, clientID string, clientSecret string) {
	log.Infof("Adding service connection to %s using client ID %s", service, clientID)

	app.gh = &oauth2.Config{
		ClientID:     clientID,
		Scopes:       []string{"repo", "read:user"},
		ClientSecret: clientSecret,
		Endpoint:     oauth2GitHub.Endpoint,
	}
}

func (app *Application) GetServiceConnection(service string) *oauth2.Config {
	return app.gh
}

func (app *Application) UpdateEpicStatus(clients *GitHubClients, event github.IssuesEvent) {
	issue := event.GetIssue()
	body := issue.GetBody()

	r, _ := regexp.Compile("^/epic #([0-9]+)")

	// we only want the command at the beginning of the line, so split it
	for _, line := range strings.Split(body, "\n") {
		match := r.FindStringSubmatch(line)

		if len(match) != 2 {
			continue
		}

		i, _ := strconv.ParseInt(match[1], 10, 64)
		epicNumber := int(i)
		epicIssueString := fmt.Sprintf("%s/%s#%d", event.Repo.Owner.GetLogin(), event.Repo.GetName(), epicNumber)

		log.Infof("Issue %s needs to be connected to epic %s", GetIssueIdentifier(event), epicIssueString)

		// find issue
		epic, _, _ := clients.V3.Issues.Get(context.Background(), event.Repo.Owner.GetLogin(), event.Repo.GetName(), epicNumber)

		epicBody := epic.GetBody()
		// the parser has problems with \r\n
		epicBody = strings.ReplaceAll(epicBody, "\r\n", "\n")
		body, status := CheckIfContainsIssue(epicBody, issue.GetTitle(), issue.GetNumber())

		if status != NotModified {
			request := github.IssueRequest{
				Body: &body,
			}

			// update issue text
			if _, _, err := clients.V3.Issues.Edit(context.Background(), event.Repo.Owner.GetLogin(), event.Repo.GetName(), epicNumber, &request); err != nil {
				log.Errorf("Updating issue %s failed: %s", epicIssueString, err)
				return
			}
		}
	}
}

func GetIssueIdentifier(event github.IssuesEvent) string {
	return fmt.Sprintf("%s/%s#%d", event.Repo.Owner.GetLogin(), event.Repo.GetName(), event.GetIssue().GetNumber())
}

type IssueUpdateStatus int

const (
	NotModified = iota
	UpdatedText
	InsertedIssue
)

func CheckIfContainsIssue(body string, title string, number int) (string, IssueUpdateStatus) {
	r := markdown.NewRenderer(&markdown.Options{})

	md := blackfriday.New()
	ast := md.Parse([]byte(body))

	log.Printf("%s", body)

	walker := &TaskWalker{IssueTitle: title, IssueNumber: number}
	ast.Walk(walker.Walk)

	var buf bytes.Buffer
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return r.RenderNode(&buf, node, entering)
	})

	return string(buf.Bytes()), walker.status
}

type TaskWalker struct {
	err         error
	currentList *blackfriday.Node
	status      IssueUpdateStatus
	exists      bool
	IssueNumber int
	IssueTitle  string
}

func (o *TaskWalker) Walk(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	//var err error

	log.Printf("%+v", node)

	if !entering && node.Type == blackfriday.List && o.currentList != nil {
		if !o.exists {
			// we did not find the issue, so we need to insert it

			text := blackfriday.NewNode(blackfriday.Text)
			text.Literal = []byte(fmt.Sprintf("[ ] %s (#%d)", o.IssueTitle, o.IssueNumber))

			p := blackfriday.NewNode(blackfriday.Paragraph)
			p.AppendChild(text)

			item := blackfriday.NewNode(blackfriday.Item)
			item.AppendChild(p)

			node.AppendChild(item)

			o.status = InsertedIssue

			// exit the walker
			return blackfriday.Terminate
		}
		o.currentList = nil
	}

	if !entering {
		return blackfriday.GoToNext
	}

	if node.Type == blackfriday.List {
		o.currentList = node
	}

	if o.currentList != nil && node.Type == blackfriday.Text {
		text := string(node.Literal)

		// inside a task list
		if strings.HasPrefix(text, "[ ]") {
			// check the text for the issue
			if strings.Contains(text, fmt.Sprintf("#%d", o.IssueNumber)) {
				o.status = NotModified
				o.exists = true
				// no need to continue
				return blackfriday.Terminate
			}
		}
	}

	return blackfriday.GoToNext
}
