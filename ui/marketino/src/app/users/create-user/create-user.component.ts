import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { UserService } from '../user.service';
import { MatDialogRef } from '@angular/material/dialog';

@Component({
  selector: 'app-create-user',
  templateUrl: './create-user.component.html',
  styleUrl: './create-user.component.css'
})
export class CreateUserComponent implements OnInit {
  form!: FormGroup;


  constructor(
    private formbuilder: FormBuilder,
    private service: UserService,
    private dialogRef: MatDialogRef<CreateUserComponent>,
  ){
   
  }

  ngOnInit(): void {
    this.form = this.formbuilder.group({
      userName: [null, Validators.required],
      firstName: [null, Validators.required],
      lastName: [null, Validators.required],
      password: [null, Validators.required],
      isAdmin: [null, Validators.required],
    })
  }

  save(){
    this.service.create(
      this.form.getRawValue()
    ).subscribe(() => {
      this.dialogRef.close();
    })
  }
}
