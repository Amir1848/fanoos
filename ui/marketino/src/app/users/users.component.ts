import { Component, ViewChild } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { ColDef } from 'ag-grid-community';
import { CreateUserComponent } from './create-user/create-user.component';
import { RemoteDataGridComponent } from '../component/remote-data-grid/remote-data-grid.component';
import { Title } from '@angular/platform-browser';
import { MarketinoTitleService } from '../services/marketino-title-service';
import { UserService } from './user.service';

@Component({
  selector: 'app-users',
  templateUrl: './users.component.html',
  styleUrl: './users.component.css',
})
export class UsersComponent {
  @ViewChild('grid') grid!: RemoteDataGridComponent;

  selectedRows: any[] = []

  colDefs: ColDef[] = [
    { field: 'firstName', headerName: 'نام', flex: 1 },
    { field: 'lastName', headerName: 'نام خانوادگی', flex: 1 },
    { field: 'userName', headerName: 'نام کاربری', flex: 1 },
    { field: 'isAdmin', headerName: 'کاربر ارشد', flex: 1 },
    { field: 'phoneNumber', headerName: 'تلفن همراه', flex: 1 },
    { field: 'email', headerName: 'ایمیل', flex: 1 },
  ];

  constructor(
    public dialog: MatDialog,
    private titleService: Title,
    private marketinoTitleService: MarketinoTitleService,
    private userService: UserService,
  ) {}

  change() {
    this.dialog
      .open(CreateUserComponent, {
        height: '400px',
        width: '800px',
      })
      .afterClosed()
      .subscribe((p) => {
        this.grid.refresh();
      });
  }

  ngOnInit(): void {
    this.titleService.setTitle('کاربران');
    this.marketinoTitleService.sendData({
      icon: 'user',
      title: 'کاربران',
    });
  }

  delete(){
    this.userService.delete(this.selectedRows.map(p => p.id)).subscribe(() => {
      this.grid.refresh()
    })
  }

  selectionChange(p: any[]){
    this.selectedRows = p

  }
}
