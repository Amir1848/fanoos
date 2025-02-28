import { Component } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { MarketinoTitleService } from '../services/marketino-title-service';

@Component({
  selector: 'app-shell',
  templateUrl: './shell.component.html',
  styleUrl: './shell.component.css'
})
export class ShellComponent {
  title = 'marketino';
  tabTitle = 'داشبورد'
  tabIcon = 'home';
  constructor(
    private titleService: Title, 
    marketinoTitleService: MarketinoTitleService
    ) {
      marketinoTitleService.data$.subscribe(p => {
        if(!!p) {
          this.tabTitle = p.title
          this.tabIcon = p.icon
        }
      })
  }
}
