import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../core/auth/auth.service';
import { AuthorizeDirective } from '../../core/auth/authorize.directive';
import { IconComponent } from '../../shared/icon/icon.component';

interface DashboardTile {
  label: string;
  desc: string;
  icon: string;
  link: string;
  color: string;
  resource?: string;
}

const TILES: DashboardTile[] = [
  {
    label: 'کاربران',
    desc: 'مدیریت کاربران سامانه، ایجاد و حذف',
    icon: 'users',
    link: '/users',
    color: 'primary',
    resource: 'users.view',
  },
  {
    label: 'نقش‌های کاربران',
    desc: 'انتخاب کاربر و مدیریت نقش‌های او',
    icon: 'link',
    link: '/users/access',
    color: 'accent',
    resource: 'user-roles.view',
  },
  {
    label: 'نقش‌ها',
    desc: 'تعریف و مدیریت نقش‌های کاربری',
    icon: 'shield',
    link: '/roles',
    color: 'accent',
    resource: 'roles.view',
  },
  {
    label: 'محیط‌ها',
    desc: 'تعریف خانه و چیدمان اتاق‌ها و دستگاه‌ها',
    icon: 'home',
    link: '/environments',
    color: 'primary',
  },
];

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [RouterLink, IconComponent, AuthorizeDirective],
  template: `
    <div class="page-header">
      <h2 class="page-title">
        داشبورد
        <span class="page-subtitle">
          خوش آمدید<span class="ltr"> {{ auth.username() }} </span>عزیز
        </span>
      </h2>
    </div>

    <div class="tiles">
      @for (tile of tiles; track tile.link) {
        <a class="tile card" *appAuthorize="tile.resource ?? ''" [routerLink]="tile.link">
          <div class="tile-icon" [class]="'tile-icon ' + tile.color">
            <app-icon [name]="tile.icon" [size]="26" />
          </div>
          <div>
            <div class="tile-title">{{ tile.label }}</div>
            <div class="tile-desc text-secondary">{{ tile.desc }}</div>
          </div>
        </a>
      }
    </div>
  `,
  styles: `
    .tiles {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
      gap: 16px;
    }

    .tile {
      display: flex;
      align-items: center;
      gap: 14px;
      padding: 18px;
      color: var(--text);
      transition: box-shadow 0.15s, transform 0.15s;

      &:hover {
        box-shadow: var(--shadow-2);
        transform: translateY(-2px);
      }
    }

    .tile-icon {
      width: 52px;
      height: 52px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;

      &.primary {
        background: var(--primary-light);
        color: var(--primary);
      }

      &.accent {
        background: #e0f2f1;
        color: var(--accent);
      }
    }

    .tile-title {
      font-weight: 700;
      font-size: 15px;
    }

    .tile-desc {
      font-size: 12px;
      margin-top: 2px;
    }
  `,
})
export class DashboardComponent {
  readonly auth = inject(AuthService);
  readonly tiles = TILES;
}
