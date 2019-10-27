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
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type RepositoryRefArray []int64

type Workspace struct {
	ID            int64              `db:"id, primarykey, autoincrement" json:"id"`
	Name          string             `db:"name" json:"name"`
	RepositoryIDs RepositoryRefArray `db:"repositoryIDs" json:"repositoryIDs"`
}

type Relationship struct {
	IssueID      int64  `db:"issueId, primarykey"`
	OtherIssueID int64  `db:"otherIssueId"`
	Type         string `db:"type"`
}

func (r *RepositoryRefArray) Scan(src interface{}) error {
	u, ok := src.([]uint8)
	if !ok {
		return errors.New("Unable to convert type from []uint8")
	}

	var intArray []int64
	var i int64
	var err error
	var s string

	s = strings.ReplaceAll(strings.ReplaceAll(string(u), "{", ""), "}", "")

	// split array
	array := strings.Split(s, ",")
	for _, v := range array {
		if i, err = strconv.ParseInt(v, 10, 64); err != nil {
			return fmt.Errorf("Could not convert all array elements to int64: %s", err)
		}

		intArray = append(intArray, i)
	}

	*r = intArray

	return nil
}

func (app *Application) GetWorkspace(workspaceID int64) (*Workspace, error) {
	var (
		workspace Workspace
		err       error
	)

	err = app.db.SelectOne(&workspace, "select * from workspace where id=$1", workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &workspace, nil
}
