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
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gopkg.in/gorp.v2"
)

type Database interface {
	Init()
	Insert(object interface{}) (err error)
	Update(object interface{}) (rowsChanged int64, err error)
	Select(holder interface{}, query string, args ...interface{}) (objects []interface{}, err error)
	SelectOne(holder interface{}, query string, args ...interface{}) (err error)
}

type MappedPostgreSQL struct {
	host   string
	db     *sql.DB
	mapper *gorp.DbMap
}

func init() {
	log = logrus.WithField("component", "db")
}

func NewMappedPostgreSQL(host string) Database {
	return &MappedPostgreSQL{host: host}
}

func (p *MappedPostgreSQL) Init() {
	p.db, _ = sql.Open("postgres", fmt.Sprintf("postgres://postgres@%s/issues?sslmode=disable", p.host))
	p.mapper = &gorp.DbMap{Db: p.db, Dialect: gorp.PostgresDialect{}}

	p.mapper.AddTableWithName(Workspace{}, "workspace").SetKeys(true, "ID")
	p.mapper.AddTableWithName(ServiceToken{}, "servicetoken").SetKeys(false, "UserID")

	log.Infof("Using PostgreSQL @ %s", p.host)
}

// Insert inserts a suitable struct into our database. Only types that are registred in InitPostgreSQL are suitable.
func (p *MappedPostgreSQL) Insert(object interface{}) (err error) {
	log.Debugf("Inserting %+v", object)

	return p.mapper.Insert(object)
}

func (p *MappedPostgreSQL) Update(object interface{}) (rowsChanged int64, err error) {
	log.Debugf("Updating %+v", object)

	return p.mapper.Update(object)
}

func (p *MappedPostgreSQL) Select(holder interface{}, query string, args ...interface{}) ([]interface{}, error) {
	return p.mapper.Select(holder, query, args...)
}

func (p *MappedPostgreSQL) SelectOne(holder interface{}, query string, args ...interface{}) error {
	return p.mapper.SelectOne(holder, query, args...)
}
