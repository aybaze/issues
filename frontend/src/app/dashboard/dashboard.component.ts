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

import { Component, OnInit } from '@angular/core';
import { WorkspaceService } from '../workspace.service';
import { Workspace } from '../workspace';
import { AuthService } from '../auth.service';
import { catchError } from 'rxjs/operators';
import { HttpErrorResponse } from '@angular/common/http';
import { EMPTY } from 'rxjs';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {

  workspaces: Workspace[];

  constructor(private workspaceService: WorkspaceService, private auth: AuthService) { }

  ngOnInit() {
    this.workspaceService.getWorkspaces()
      .pipe(
        catchError((err: HttpErrorResponse) => {
          if (err.status === 401) {
            this.auth.logout();
          }
          return EMPTY;
        }))
      .subscribe(workspaces => {
        this.workspaces = workspaces;
      });
  }

}
