import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { WorkspaceService } from '../workspace.service';
import { Workspace } from '../workspace';

@Component({
  selector: 'app-workspace-detail',
  templateUrl: './workspace-detail.component.html',
  styleUrls: ['./workspace-detail.component.scss']
})
export class WorkspaceDetailComponent implements OnInit {

  workspaceId: number;
  workspace: Workspace;

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
    })
  }

}
