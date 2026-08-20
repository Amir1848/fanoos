import { Component, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { AuthService } from '../../core/auth/auth.service';
import { AuthorizeDirective } from '../../core/auth/authorize.directive';
import { ConfirmService } from '../../shared/confirm/confirm.service';
import { IconComponent } from '../../shared/icon/icon.component';

const SIDEBAR_KEY = 'fanoos.sidebar-collapsed';

interface NavItem {
  label: string;
  icon: string;
  link: string;
  resource?: string;
}

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, IconComponent, AuthorizeDirective],
  templateUrl: './shell.component.html',
  styleUrl: './shell.component.scss',
})
export class ShellComponent {
  readonly auth = inject(AuthService);
  private readonly confirm = inject(ConfirmService);

  readonly collapsed = signal(localStorage.getItem(SIDEBAR_KEY) === '1');

  readonly navItems: NavItem[] = [
    { label: 'داشبورد', icon: 'dashboard', link: '/dashboard' },
    { label: 'کاربران', icon: 'users', link: '/users', resource: 'users.view' },
    { label: 'نقش‌های کاربران', icon: 'link', link: '/users/access', resource: 'user-roles.view' },
    { label: 'نقش‌ها', icon: 'shield', link: '/roles', resource: 'roles.view' },
    { label: 'محیط‌ها', icon: 'home', link: '/environments' },
  ];

  toggleSidebar(): void {
    this.collapsed.update((v) => !v);
    localStorage.setItem(SIDEBAR_KEY, this.collapsed() ? '1' : '0');
  }

  async logout(): Promise<void> {
    const ok = await this.confirm.ask({
      title: 'خروج از سامانه',
      message: 'آیا می‌خواهید از حساب کاربری خود خارج شوید؟',
      confirmText: 'خروج',
    });

    if (ok) {
      this.auth.logout();
    }
  }
}
