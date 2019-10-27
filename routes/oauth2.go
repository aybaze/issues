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
	"fmt"
	"issues"
	"net/http"
	"time"

	"github.com/google/go-github/github"
	"github.com/oxisto/go-httputil/auth"
	"golang.org/x/oauth2"
	gh "golang.org/x/oauth2/github"
)

var (
	ctx       context.Context
	conf      *oauth2.Config
	jwtSecret string
)

func init() {
	ctx = context.Background()
}

// AddServiceConnection adds a new connection to an external service. For now this
// just statically sets up the GitHub connection
func AddServiceConnection(service string, clientID string, clientSecret string) {
	log.Infof("Adding service connection to %s using client ID %s", service, clientID)

	conf = &oauth2.Config{
		ClientID:     clientID,
		Scopes:       []string{"repo", "read:user"},
		ClientSecret: clientSecret,
		Endpoint:     gh.Endpoint,
	}
}

// SetJWTSecret sets the JWT secret used for signing tokens issued by our API
func SetJWTSecret(secret string) {
	jwtSecret = secret
}

func handleOAuthFlowError(err error, w http.ResponseWriter, r *http.Request) {
	log.Errorf("Could not fetch access token: %v", err)
	w.Header().Add("Location", "/oauth2/login")
	w.WriteHeader(http.StatusFound)
}

func handleOAuth2Login(w http.ResponseWriter, r *http.Request) {
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOnline)

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusFound)
}

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	var (
		serviceToken *oauth2.Token
		apiToken     *oauth2.Token
		user         *github.User
		err          error
	)
	code := r.URL.Query().Get("code")

	log.Infof("Got callback for code %s", code)

	// fetch access token with authorization code
	serviceToken, err = conf.Exchange(ctx, code)
	if err != nil {
		handleOAuthFlowError(err, w, r)
		return
	}

	tc := conf.Client(ctx, serviceToken)

	// create a GitHub client
	gc := github.NewClient(tc)

	if user, _, err = gc.Users.Get(ctx, ""); err != nil {
		handleOAuthFlowError(err, w, r)
		return
	}

	// cache the service token
	err = issues.AddServiceToken(&issues.ServiceToken{
		user.GetID(),
		issues.ServiceGitHub,
		serviceToken.AccessToken})
	if err != nil {
		// no chance to recover
		log.Errorf("Could not add service token to database: %s", err)
		w.WriteHeader(500)
		return
	}

	// go on and issue an API token for ourselves
	// issue an authentication token for our own API
	apiToken, err = auth.IssueToken([]byte(jwtSecret), fmt.Sprintf("%d", user.GetID()), time.Now().Add(1*time.Hour))
	if err != nil {
		handleOAuthFlowError(err, w, r)
		return
	}

	// redirect to main frontend page
	w.Header().Add("Location", "/#?token="+apiToken.AccessToken+"&github_token="+serviceToken.AccessToken)
	w.Header().Add("Set-Cookie", "token="+apiToken.AccessToken+";github_token="+serviceToken.AccessToken+" Path=/")
	w.WriteHeader(http.StatusFound)
}
