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

func (router *Router) handleGetWorkspaces(w http.ResponseWriter, r *http.Request) {
	workspaces, err := router.app.GetDatabase().GetWorkspaces(nil)

	httputil.JSONResponse(w, r, workspaces, err)
}

func (router *Router) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
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
	workspace, err = router.app.GetWorkspace(workspaceID)
	httputil.JSONResponse(w, r, workspace, err)

	return
}

func (router *Router) handleGetIssues(w http.ResponseWriter, r *http.Request) {
	var (
		workspaceID int64
		err         error
	)
	if workspaceID, err = strconv.ParseInt(mux.Vars(r)["workspaceID"], 10, 64); err != nil {
		httputil.JSONResponse(w, r, nil, err)
		return
	}

	clients := r.Context().Value(issues.ServiceGitHub).(*issues.GitHubClients)

	issues, err := router.app.GetBacklog(clients, workspaceID)
	// TODO: check somehow, if user has access

	httputil.JSONResponse(w, r, issues, err)
}
