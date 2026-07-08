import { Component, inject } from '@angular/core';
import { ConfirmService } from './confirm.service';

@Component({
  selector: 'app-confirm-dialog',
  standalone: true,
  template: `
    @if (confirmService.pending(); as pending) {
      <div class="overlay" (click)="confirmService.answer(false)">
        <div class="dialog card" role="dialog" aria-modal="true" (click)="$event.stopPropagation()">
          <h3 class="dialog-title">{{ pending.title }}</h3>
          <p class="dialog-message">{{ pending.message }}</p>
          <div class="dialog-actions">
            <button type="button" class="btn btn-outline" (click)="confirmService.answer(false)">
              {{ pending.cancelText || 'انصراف' }}
            </button>
            <button
              type="button"
              class="btn"
              [class.btn-danger]="pending.danger"
              [class.btn-primary]="!pending.danger"
              (click)="confirmService.answer(true)"
            >
              {{ pending.confirmText || 'تأیید' }}
            </button>
          </div>
        </div>
      </div>
    }
  `,
  styles: `
    .overlay {
      position: fixed;
      inset: 0;
      background: rgba(20, 24, 40, 0.45);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 1000;
      animation: fade-in 0.15s ease;
    }

    .dialog {
      width: min(420px, calc(100vw - 32px));
      padding: 20px;
      animation: pop-in 0.18s ease;
    }

    .dialog-title {
      font-size: 15px;
      margin-bottom: 8px;
    }

    .dialog-message {
      margin: 0 0 20px;
      color: var(--text-secondary);
      font-size: 13px;
    }

    .dialog-actions {
      display: flex;
      justify-content: flex-end;
      gap: 8px;
    }

    @keyframes fade-in {
      from { opacity: 0; }
      to { opacity: 1; }
    }

    @keyframes pop-in {
      from { opacity: 0; transform: scale(0.96); }
      to { opacity: 1; transform: scale(1); }
    }
  `,
})
export class ConfirmDialogComponent {
  readonly confirmService = inject(ConfirmService);
}
