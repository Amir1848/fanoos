import { Component, ViewChild } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { ColDef } from 'ag-grid-community';
import { CreateUserComponent } from './create-user/create-user.component';
import { RemoteDataGridComponent } from '../component/remote-data-grid/remote-data-grid.component';



@Component({
  selector: 'app-users',
  templateUrl: './users.component.html',
  styleUrl: './users.component.css'
})
export class UsersComponent {

  @ViewChild('grid') grid!: RemoteDataGridComponent;

  colDefs: ColDef[] = [
    { field: "FirstName", flex: 1 },
    { field: "LastName", flex: 1 }
  ];

  constructor(
    public dialog: MatDialog){

  }

  change(){
    this.dialog.open(CreateUserComponent, {
      height: '400px',
      width: '800px',
    }).afterClosed().subscribe(p => {
      this.grid.refresh()
    });

    
  }
}
