import { Component, EventEmitter, Input, Output, OnInit, OnDestroy } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';

@Component({
  templateUrl: './login.component.html',
  styleUrl: './login.component.scss'
})
export class LoginComponent  implements OnInit, OnDestroy {
  form: FormGroup = new FormGroup({
    username: new FormControl(''),
    password: new FormControl(''),
  });

  submit() {
    if (this.form.valid) {
      this.submitEM.emit(this.form.value);
    }
  }
  @Input() error!: string | null;

  @Output() submitEM = new EventEmitter();

  ngOnInit() {
    document.body.classList.add('login-background');
  }

  ngOnDestroy() {
    document.body.classList.remove('login-background');
  }
}