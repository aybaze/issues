import { Component, OnInit, Input } from '@angular/core';

@Component({
  selector: 'issues-user-link',
  templateUrl: './user-link.component.html',
  styleUrls: ['./user-link.component.scss']
})
export class UserLinkComponent implements OnInit {

  @Input() user: any;

  constructor() { }

  ngOnInit() {
  }

}
