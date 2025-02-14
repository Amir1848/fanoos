import { Component } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { ColDef } from 'ag-grid-community';
import { CreateUserComponent } from './create-user/create-user.component';



@Component({
  selector: 'app-users',
  templateUrl: './users.component.html',
  styleUrl: './users.component.css'
})
export class UsersComponent {
  colDefs: ColDef[] = [
    { field: "FirstName", flex: 1 },
    { field: "LastName", flex: 1 }
  ];

  constructor(public dialog: MatDialog){

  }

  change(){
    let dialogRef = this.dialog.open(CreateUserComponent, {
      height: '400px',
      width: '800px',
    });

    
  }
}
