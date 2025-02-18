import { Component } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { MarketinoTitleService } from './services/marketino-title-service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})
export class AppComponent {
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
