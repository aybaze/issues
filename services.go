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
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v29/github"
	"github.com/gregjones/httpcache"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// TokenCache contains a simple cache of github clients
var clients map[int64]*GitHubClients
var installationClients map[int64]*GitHubClients

var ErrAuthenticationNeeded = errors.New("You need to authenticate with the service")

const ServiceGitHub = "GitHub"

// GitHubClients is a structure that provides v3 (REST) and v4 (GraphQL) GitHub clients
type GitHubClients struct {
	// The authenticated user, if this is a user-client
	User *github.User
	// Specifices, if this is a user client or installation client
	IsUserClient bool
	V3           *github.Client
	V4           *githubv4.Client
}

// ServiceToken represents an OAuth2-style token to an external service, such as GitHub
type ServiceToken struct {
	UserID      int64  `db:"userId, primarykey"`
	Service     string `db:"service"`
	AccessToken string `db:"accessToken"`
}

func (app *Application) newGitHubClients(userID int64) (clients *GitHubClients, err error) {
	// fetch token from db
	var serviceToken ServiceToken
	if err = app.db.SelectOne(&serviceToken, "select * from servicetoken where service=$1 and \"userId\"=$2", "GitHub", userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuthenticationNeeded
		}

		return nil, fmt.Errorf("Could not fetch GitHub token from database: %w", err)
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: serviceToken.AccessToken,
	})

	cachedTransport := httpcache.NewMemoryCacheTransport()
	cachedTransport.Transport = &oauth2.Transport{
		Source: tokenSource,
	}

	httpClient := &http.Client{Transport: cachedTransport}

	// creating new GitHub clients
	clients = &GitHubClients{
		IsUserClient: true,
		V3:           github.NewClient(httpClient),
		V4:           githubv4.NewClient(httpClient),
	}

	// fetch a user, to test the client
	if clients.User, _, err = clients.V3.Users.Get(context.Background(), ""); err != nil {
		return nil, fmt.Errorf("Could not create GitHub clients because authenticated user could not be retrieved: %s", err)
	}

	log.Debugf("Succesfully created GitHub clients for authenticated user %s", clients.User.GetLogin())

	return clients, nil
}

func (app *Application) newGitHubInstallationClients(installationID int64) (clients *GitHubClients, err error) {
	tr := http.DefaultTransport

	itr, err := ghinstallation.NewKeyFromFile(tr, app.AppID, installationID, "keys/private-key.pem")
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Transport: itr}

	// creating new GitHub clients
	clients = &GitHubClients{
		IsUserClient: false,
		V3:           github.NewClient(httpClient),
		V4:           githubv4.NewClient(httpClient),
	}

	log.Debugf("Succesfully created GitHub clients for installation ID %d", installationID)

	return
}

func init() {
	clients = make(map[int64]*GitHubClients)
	installationClients = make(map[int64]*GitHubClients)
}

func (app *Application) AddServiceToken(token *ServiceToken) (err error) {
	var old *ServiceToken

	// check, if it already exists
	old, err = app.GetServiceToken(token.Service, token.UserID)
	if err != nil {
		return err
	}

	if old == nil {
		return app.db.Insert(token)
	}

	old.AccessToken = token.AccessToken

	_, err = app.db.Update(token)

	// force in-memory cache to refresh
	if token.Service == ServiceGitHub {
		delete(clients, token.UserID)
	}

	return
}

func (app *Application) GetServiceToken(service string, userID int64) (token *ServiceToken, err error) {
	var t ServiceToken
	err = app.db.SelectOne(&t, "select * from servicetoken where service=$1 and \"userId\"=$2", service, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	token = &t

	return token, err
}

func (app *Application) GetUserClients(userID int64) (c *GitHubClients, err error) {
	var (
		found bool
	)

	c, found = clients[userID]

	if found {
		log.Debugf("Using in-memory GitHub clients for authenticated user %s", c.User.GetLogin())
		return c, nil
	}

	if c, err = app.newGitHubClients(userID); err != nil {
		return nil, err
	}

	clients[userID] = c
	return
}

func (app *Application) GetInstallationClients(installationID int64) (c *GitHubClients, err error) {
	var (
		found bool
	)

	c, found = installationClients[installationID]

	if found {
		log.Debugf("Using in-memory GitHub clients for installation %d", installationID)
		return c, nil
	}

	if c, err = app.newGitHubInstallationClients(installationID); err != nil {
		return nil, err
	}

	installationClients[installationID] = c
	return
}
