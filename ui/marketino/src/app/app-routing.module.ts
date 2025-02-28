import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { UsersComponent } from './users/users.component';
import { GeneralMenuComponent } from './general/general-menu/general-menu.component';
import { LoginComponent } from './login/login.component';
import { ShellComponent } from './shell/shell.component';

const routes: Routes = [
  {
    path: 'login',
    component: LoginComponent
  },
  {
    path: '',
    component: ShellComponent,
    children: [
      {
        path: 'general',
        component: GeneralMenuComponent
      },
      {
        path: 'general/users',
        component: UsersComponent
      }
    ]
  }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
