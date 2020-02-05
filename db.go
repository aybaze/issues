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
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
)

type Database interface {
	Init()
	Insert(object interface{}) (err error)
	Update(object interface{}) (rowsChanged int64, err error)
	GetServiceToken(service string, userID int64) (*ServiceToken, error)
	GetWorkspace(workspaceID int64) (*Workspace, error)
	GetWorkspaces(query interface{}, args ...interface{}) ([]*Workspace, error)
	GetRelationships(query interface{}, args ...interface{}) ([]*Relationship, error)
}

type MappedPostgreSQL struct {
	host string
	db   *gorm.DB
}

func init() {
	log = logrus.WithField("component", "db")
}

func NewMappedPostgreSQL(host string) Database {
	return &MappedPostgreSQL{host: host}
}

func (p *MappedPostgreSQL) Init() {
	var err error

	if p.db, err = gorm.Open("postgres", fmt.Sprintf("postgres://postgres@%s/issues?sslmode=disable", p.host)); err != nil {
		panic(err)
	}

	p.db.AutoMigrate(&Workspace{})
	p.db.AutoMigrate(&ServiceToken{})
	p.db.AutoMigrate(&Relationship{})

	log.Infof("Using PostgreSQL @ %s", p.host)
}

// Insert inserts a suitable struct into our database. Only types that are registred in InitPostgreSQL are suitable.
func (p *MappedPostgreSQL) Insert(object interface{}) (err error) {
	log.Debugf("Inserting %+v", object)

	return p.db.Create(object).Error
}

func (p *MappedPostgreSQL) Update(object interface{}) (rowsChanged int64, err error) {
	log.Debugf("Updating %+v", object)

	scoped := p.db.Save(object)
	rowsChanged = scoped.RowsAffected
	err = scoped.Error
	return
}

func (p *MappedPostgreSQL) Where(holder interface{}, query string, args ...interface{}) error {
	return p.db.Where(query, args).Find(&holder).Error
}

func (p *MappedPostgreSQL) GetWorkspace(workspaceID int64) (*Workspace, error) {
	var w Workspace
	err := p.db.First(&w, workspaceID).Error

	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &w, err
}

func (p *MappedPostgreSQL) GetWorkspaces(query interface{}, args ...interface{}) ([]*Workspace, error) {
	var (
		w   []*Workspace
		err error
		db  *gorm.DB
	)

	db = p.db

	if query != nil {
		db = db.Where(query, args)
	}

	err = db.Find(&w).Error

	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return w, err
}

func (p *MappedPostgreSQL) GetServiceToken(service string, userID int64) (*ServiceToken, error) {
	t := ServiceToken{
		Service: "GitHub",
		UserID:  userID,
	}

	var err error

	err = p.db.Where(&t).First(&t).Error

	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return &t, err
}

func (p *MappedPostgreSQL) GetRelationships(query interface{}, args ...interface{}) ([]*Relationship, error) {
	var (
		r   []*Relationship
		err error
		db  *gorm.DB
	)

	db = p.db

	if query != nil {
		db = db.Where(query, args)
	}

	err = db.Find(&r).Error

	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}

	return r, err
}
