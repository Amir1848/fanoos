import { Component, OnInit } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { MarketinoTitleService } from '../services/marketino-title-service';

@Component({
  selector: 'app-general-menu',
  templateUrl: './general-menu.component.html',
  styleUrl: './general-menu.component.css'
})
export class GeneralMenuComponent implements OnInit {
  constructor(
    private titleService: Title,
    private marketinoTitleService: MarketinoTitleService,
  ){
    
  }

  ngOnInit(): void {
    this.titleService.setTitle("عمومی")
    this.marketinoTitleService.sendData({
      icon: 'desktop',
      title: "عمومی"
    })
  }
  
}
