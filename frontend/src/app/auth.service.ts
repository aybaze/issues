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

import { Injectable } from '@angular/core';
import { JwtHelperService } from '@auth0/angular-jwt';
import { Router, ActivatedRoute } from '@angular/router';
import { HttpParams } from '@angular/common/http';

export const TOKEN = 'token';
export const GITHUB_TOKEN = 'github_token';

const helper = new JwtHelperService();

@Injectable({
  providedIn: 'root'
})
export class AuthService {

  constructor(private router: Router, private route: ActivatedRoute) {
    const params = new HttpParams({ fromString: window.location.hash.replace('#?', '') });

    const token = params.get('token');
    const gitHubToken = params.get('github_token');

    if (token && gitHubToken) {
      this.login(token, gitHubToken);
      this.router.navigate(['/']);
    }
  }

  login(token: string, gitHubToken: string) {
    localStorage.setItem(TOKEN, token);
    localStorage.setItem(GITHUB_TOKEN, gitHubToken);
  }

  logout() {
    localStorage.removeItem(TOKEN);
    localStorage.removeItem(GITHUB_TOKEN);

    this.router.navigateByUrl('/login');
  }

  getToken() {
    return localStorage.getItem(TOKEN);
  }

  getGitHubToken() {
    return localStorage.getItem(GITHUB_TOKEN);
  }

  isLoggedIn(): boolean {
    return !helper.isTokenExpired(this.getToken()) && this.getGitHubToken() !== undefined;
  }
}
