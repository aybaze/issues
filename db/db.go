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

package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gopkg.in/gorp.v2"
)

var pdb *sql.DB
var log *logrus.Entry
var mapper *gorp.DbMap

func init() {
	log = logrus.WithField("component", "db")
}

func InitPostgreSQL(host string) {
	pdb, _ = sql.Open("postgres", fmt.Sprintf("postgres://postgres@%s/postgres?sslmode=disable", host))
	mapper = &gorp.DbMap{Db: pdb, Dialect: gorp.PostgresDialect{}}

	log.Infof("Using PostgreSQL @ %s", host)
}

func GetMapper() *gorp.DbMap {
	return mapper
}

// Insert inserts a suitable struct into our database. Only types that are registred in InitPostgreSQL are suitable.
func Insert(object interface{}) (err error) {
	log.Debugf("Inserting %+v", object)

	return mapper.Insert(object)
}

func Select(i interface{}, query string, args ...interface{}) ([]interface{}, error) {
	return mapper.Select(i, query, args...)
}

func SelectOne(holder interface{}, query string, args ...interface{}) error {
	return mapper.SelectOne(holder, query, args...)
}
