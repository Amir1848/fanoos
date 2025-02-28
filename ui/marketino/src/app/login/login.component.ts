import { Component, EventEmitter, Input, Output, OnInit, OnDestroy } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { loginService } from './login.service';
import { Router } from '@angular/router';

@Component({
  templateUrl: './login.component.html',
  styleUrl: './login.component.scss'
})
export class LoginComponent  implements OnInit, OnDestroy {
  form: FormGroup = new FormGroup({
    username: new FormControl('', [Validators.required]),
    password: new FormControl('', [Validators.required]),
  });

  constructor(
    private loginService: loginService,
    private router: Router
  ){

  }

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

  login(){
    if (this.form.valid) {
      this.loginService.login(this.form.getRawValue()).then(result => {
        this.error = ""
        localStorage.setItem('token', result.token);
        this.router.navigateByUrl('/')
      }).catch(() => {
        this.error = "رمز عبور و نام کاربری صحیح نیست"
      })
    } else {
      this.error = "نام کاربری و رمز عبور را وارد کنید"
    }
  }
}