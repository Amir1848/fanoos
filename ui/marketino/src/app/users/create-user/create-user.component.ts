import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

@Component({
  selector: 'app-create-user',
  templateUrl: './create-user.component.html',
  styleUrl: './create-user.component.css'
})
export class CreateUserComponent {
  form: FormGroup

  constructor(private formbuilder: FormBuilder){
    this.form = formbuilder.group({
      userName: [null, Validators.required]
    })
  }
}
