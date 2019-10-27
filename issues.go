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
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
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
		Endpoint:     github.Endpoint,
	}
}

func (app *Application) GetServiceConnection(service string) *oauth2.Config {
	return app.gh
}
