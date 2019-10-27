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
	"errors"
	"issues"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/oxisto/go-httputil/auth"
	"github.com/urfave/negroni"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Entry
)

func init() {
	log = logrus.WithField("component", "routes")
}

type Router struct {
	*mux.Router
	app *issues.Application
}

func (router *Router) WithMiddleware(handler *auth.JWTHandler, handlerFunc http.HandlerFunc) *negroni.Negroni {
	return negroni.New(
		negroni.HandlerFunc(handler.HandleWithNext),
		negroni.HandlerFunc(router.HandleDecodeClaimsWithNext),
		negroni.Wrap(handlerFunc),
	)
}

func HandleError(err error, w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	log.Errorf("An error occured in the processing chain: %s", err)

	var ve *jwt.ValidationError
	if errors.As(err, &ve) {
		// invalid JWT
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// otherwise, we cannot process it
	w.WriteHeader(http.StatusInternalServerError)
	return
}

// HandleDecodeClaimsWithNext extracts the claims from the supplied JWT and injects
// it into the request context
func (router *Router) HandleDecodeClaimsWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var (
		token   *jwt.Token
		claims  *jwt.StandardClaims
		userID  int64
		service string
		clients *issues.GitHubClients
		err     error
		ok      bool
	)

	app := router.app
	token, ok = r.Context().Value(auth.DefaultAuthContext).(*jwt.Token)
	if !ok {
		log.Errorf("Got invalid claims")
		w.WriteHeader(401)
		return
	}

	claims, ok = token.Claims.(*jwt.StandardClaims)
	if !ok {
		log.Errorf("Got invalid claims")
		w.WriteHeader(401)
		return
	}

	userID, err = strconv.ParseInt(claims.Subject, 10, 64)
	if !ok {
		log.Errorf("Could not parse user id from claims: %s", claims.Subject)
		w.WriteHeader(401)
		return
	}

	service = issues.ServiceGitHub

	if clients, err = app.GetUserClients(userID); err != nil {
		if errors.Is(err, issues.ErrAuthenticationNeeded) {
			log.Errorf("Could not find valid token for service %s: %s", service, err)
			w.WriteHeader(401)
			return
		}

		log.Errorf("An error occured while creating clients for service %s: %s", service, err)
		w.WriteHeader(500)
		return
	}

	request := r.WithContext(context.WithValue(r.Context(), service, clients))

	*r = *request
	next(w, r)
}

func NewRouter(app *issues.Application, jwtSecret string) *Router {
	// set the JWT secret so its accessible in API handlers
	SetJWTSecret(jwtSecret)

	options := auth.DefaultOptions
	options.JWTKeySupplier = func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	}
	options.TokenExtractor = auth.ExtractFromFirstAvailable(
		auth.ExtractTokenFromCookie("auth"),
		auth.ExtractTokenFromHeader)
	options.ErrorHandler = HandleError
	handler := auth.NewHandler(options)

	router := &Router{mux.NewRouter().StrictSlash(true), app}
	router.HandleFunc("/oauth2/callback", router.handleOAuth2Callback)
	router.HandleFunc("/oauth2/login", router.handleOAuth2Login)
	router.HandleFunc("/github/callback", router.handleGitHubCallback).Methods("POST")
	router.Handle("/api/v1/workspaces/", router.WithMiddleware(handler, router.handleGetWorkspaces)).Methods("GET")
	router.Handle("/api/v1/workspaces/{workspaceID}", router.WithMiddleware(handler, router.handleGetWorkspace)).Methods("GET")
	router.Handle("/api/v1/workspaces/{workspaceID}/issues", router.WithMiddleware(handler, router.handleGetIssues)).Methods("GET")

	return router
}
