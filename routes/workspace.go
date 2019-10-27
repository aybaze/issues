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
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/oxisto/go-httputil"
)

func handleGetWorkspaces(w http.ResponseWriter, r *http.Request) {
	var workspaces []issues.Workspace
	_, err := issues.GetDatabase().Select(&workspaces, "select * from workspace")

	httputil.JSONResponse(w, r, workspaces, err)
}

func handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	var (
		workspaceID int64
		workspace   *issues.Workspace
		err         error
	)

	if workspaceID, err = strconv.ParseInt(mux.Vars(r)["workspaceID"], 10, 64); err != nil {
		httputil.JSONResponse(w, r, nil, err)
		return
	}

	// TODO: check somehow, if user has access
	workspace, err = issues.GetWorkspace(workspaceID, issues.GetDatabase())
	httputil.JSONResponse(w, r, workspace, err)

	return
}

func handleGetIssues(w http.ResponseWriter, r *http.Request) {
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
