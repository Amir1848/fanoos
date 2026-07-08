import { Component, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../core/auth/auth.service';
import { IconComponent } from '../../shared/icon/icon.component';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [RouterLink, IconComponent],
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
      <a class="tile card" routerLink="/users">
        <div class="tile-icon users-icon">
          <app-icon name="users" [size]="26" />
        </div>
        <div>
          <div class="tile-title">کاربران</div>
          <div class="tile-desc text-secondary">مدیریت کاربران سامانه، ایجاد و حذف</div>
        </div>
      </a>

      <a class="tile card" routerLink="/roles">
        <div class="tile-icon roles-icon">
          <app-icon name="shield" [size]="26" />
        </div>
        <div>
          <div class="tile-title">نقش‌ها</div>
          <div class="tile-desc text-secondary">تعریف و مدیریت نقش‌های کاربری</div>
        </div>
      </a>
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
    }

    .users-icon {
      background: var(--primary-light);
      color: var(--primary);
    }

    .roles-icon {
      background: #e0f2f1;
      color: var(--accent);
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
}
