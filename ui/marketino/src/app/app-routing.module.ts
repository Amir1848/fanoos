import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { UsersComponent } from './users/users.component';
import { GeneralMenuComponent } from './general/general-menu/general-menu.component';

const routes: Routes = [
  {
    path: 'general',
    component: GeneralMenuComponent
  },
  {
    path: 'general/users',
    component: UsersComponent
  }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
