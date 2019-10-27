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

import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Workspace } from './workspace';
import { map } from 'rxjs/operators';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class WorkspaceService {

  constructor(private http: HttpClient) {

  }

  getWorkspaces(): Observable<Workspace[]> {
    return this.http.get<any[]>('/api/v1/workspaces').pipe(map(entries => {
      return entries.map(entry => Object.assign(new Workspace(), entry));
    }));;
  }

  getWorkspace(workspaceID: number): Observable<Workspace> {
    return this.http.get<any[]>('/api/v1/workspace/' + workspaceID).pipe(map(data => {
      return Object.assign(new Workspace(), data)
    }));
  }
}
