import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { WorkspaceService } from '../workspace.service';
import { Workspace, Issue } from '../workspace';

@Component({
  selector: 'issues-workspace-detail',
  templateUrl: './workspace-detail.component.html',
  styleUrls: ['./workspace-detail.component.scss']
})
export class WorkspaceDetailComponent implements OnInit {

  workspaceId: number;
  workspace: Workspace;
  backlog: Issue[];

  constructor(private workspaceService: WorkspaceService, private route: ActivatedRoute) {
  }

  ngOnInit() {
    this.route.params.subscribe(params => {
      this.workspaceId = params['id'];

      this.updateWorkspace();
    });
  }

  updateWorkspace() {
    this.workspaceService.getWorkspace(this.workspaceId).subscribe(workspace => {
      this.workspace = workspace;

      this.fetchBacklog();
    });
  }

  fetchBacklog() {
    this.workspaceService.getIssuers(this.workspaceId).subscribe(issues => {
      this.backlog = issues;
    });
  }

}
