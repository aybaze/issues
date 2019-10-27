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

package main

import (
	"issues"
	"issues/routes"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/oxisto/go-httputil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	PostgresFlag              = "postgres"
	ListenFlag                = "listen"
	JwtSecretFlag             = "jwt.secret"
	GitHubAppIDFlag           = "github.app.id"
	GitHubAppClientIDFlag     = "github.app.clientID"
	GitHubAppClientSecretFlag = "github.app.clientSecret"

	DefaultPostgres = "localhost"
	DefaultListen   = ":8000"
	DefaultEmpty    = ""

	EnvPrefix = "ISSUES"
)

var cmd = &cobra.Command{
	Use:   "issues",
	Short: "issuesis the main API server for Issues",
	Long:  "This is the main component of Issues. It serves as a GitHub App to manage issues.",
	Run:   doCmd,
}

func init() {
	cobra.OnInitialize(initConfig)

	cmd.Flags().String(ListenFlag, DefaultListen, "Host and port to listen to")
	cmd.Flags().String(PostgresFlag, DefaultPostgres, "Connection string for PostgreSQL")
	cmd.Flags().String(JwtSecretFlag, DefaultEmpty, "The secret used for signing API tokens")
	cmd.Flags().String(GitHubAppIDFlag, DefaultEmpty, "The GitHub App ID")
	cmd.Flags().String(GitHubAppClientIDFlag, DefaultEmpty, "The GitHub App Client ID")
	cmd.Flags().String(GitHubAppClientSecretFlag, DefaultEmpty, "The GitHub App ID Client Secret")

	viper.BindPFlag(ListenFlag, cmd.Flags().Lookup(ListenFlag))
	viper.BindPFlag(PostgresFlag, cmd.Flags().Lookup(PostgresFlag))
	viper.BindPFlag(JwtSecretFlag, cmd.Flags().Lookup(JwtSecretFlag))
	viper.BindPFlag(GitHubAppIDFlag, cmd.Flags().Lookup(GitHubAppIDFlag))
	viper.BindPFlag(GitHubAppClientIDFlag, cmd.Flags().Lookup(GitHubAppClientIDFlag))
	viper.BindPFlag(GitHubAppClientSecretFlag, cmd.Flags().Lookup(GitHubAppClientSecretFlag))
}

func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")

	// TODO: should we read config here ?!
	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("Could not read config: %s", err)
	}
}

func doCmd(cmd *cobra.Command, args []string) {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.Info("Starting server...")

	log.SetLevel(log.DebugLevel)

	db := issues.NewMappedPostgreSQL(viper.GetString(PostgresFlag))
	appID := viper.GetInt64(GitHubAppIDFlag)

	app := issues.NewApplication(appID, db)
	app.AddServiceConnection(issues.ServiceGitHub, viper.GetString(GitHubAppClientIDFlag), viper.GetString(GitHubAppClientSecretFlag))
	router := handlers.LoggingHandler(&httputil.LogWriter{Level: log.DebugLevel, Component: "http"}, routes.NewRouter(app, viper.GetString(JwtSecretFlag)))

	listen := viper.GetString(ListenFlag)

	log.Infof("Starting API on %s", listen)

	err := http.ListenAndServe(listen, router)

	log.Errorf("An error occured: %v", err)
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
