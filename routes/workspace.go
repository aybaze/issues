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
	"issues"
	"issues/db"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/oxisto/go-httputil"
)

func GetWorkspaces(w http.ResponseWriter, r *http.Request) {
	var workspaces []issues.Workspace
	_, err := db.Select(&workspaces, "select * from workspace")

	httputil.JSONResponse(w, r, workspaces, err)
}

func GetWorkspace(w http.ResponseWriter, r *http.Request) {
	var (
		workspaceID int
		err         error
	)
	if workspaceID, err = strconv.Atoi(mux.Vars(r)["workspaceID"]); err != nil {
		httputil.JSONResponse(w, r, nil, err)
		return
	}

	log.Infof("Fetching workspace %d", workspaceID)

	// TODO: check somehow, if user has access
}

func GetIssues(w http.ResponseWriter, r *http.Request) {
	var (
		workspaceID int
		err         error
	)
	if workspaceID, err = strconv.Atoi(mux.Vars(r)["workspaceID"]); err != nil {
		httputil.JSONResponse(w, r, nil, err)
		return
	}

	log.Infof("Fetching workspace %d", workspaceID)

	// TODO: check somehow, if user has access
}
