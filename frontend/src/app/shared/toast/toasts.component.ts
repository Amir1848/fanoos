import { Component, inject } from '@angular/core';
import { ToastService } from './toast.service';

@Component({
  selector: 'app-toasts',
  standalone: true,
  template: `
    <div class="toast-stack" role="status" aria-live="polite">
      @for (toast of toastService.toasts(); track toast.id) {
        <div class="toast" [class]="'toast toast-' + toast.kind">
          <span class="toast-message">{{ toast.message }}</span>
          <button type="button" class="toast-close" (click)="toastService.dismiss(toast.id)" aria-label="بستن">
            ✕
          </button>
        </div>
      }
    </div>
  `,
  styles: `
    .toast-stack {
      position: fixed;
      bottom: 20px;
      inset-inline-start: 20px;
      display: flex;
      flex-direction: column;
      gap: 8px;
      z-index: 1100;
      max-width: 360px;
    }

    .toast {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 10px 14px;
      border-radius: 8px;
      color: #fff;
      box-shadow: 0 4px 16px rgba(20, 24, 40, 0.25);
      animation: toast-in 0.25s ease;
      font-size: 13px;
    }

    .toast-success { background: #2e7d32; }
    .toast-error { background: #c62828; }
    .toast-info { background: #33409e; }

    .toast-message { flex: 1; }

    .toast-close {
      background: none;
      border: none;
      color: rgba(255, 255, 255, 0.8);
      cursor: pointer;
      font-size: 12px;
      padding: 2px;

      &:hover { color: #fff; }
    }

    @keyframes toast-in {
      from { opacity: 0; transform: translateY(8px); }
      to { opacity: 1; transform: translateY(0); }
    }
  `,
})
export class ToastsComponent {
  readonly toastService = inject(ToastService);
}
