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
  rowData = [
    { make: "Tesla", model: "Model Y", price: 64950, electric: true },
    { make: "Ford", model: "F-Series", price: 33850, electric: false },
    { make: "Toyota", model: "Corolla", price: 29600, electric: false },
  ];
 
  colDefs: ColDef[] = [
    { field: "make", flex: 1 },
    { field: "model", flex: 1 },
    { field: "price", flex: 1 },
    { field: "electric", flex: 1 }
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
