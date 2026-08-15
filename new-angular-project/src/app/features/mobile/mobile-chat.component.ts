import { Component, DestroyRef, ElementRef, effect, inject, signal, viewChild } from '@angular/core';
import { Router } from '@angular/router';
import { AssistantService } from '../../core/api/assistant.service';
import { ChatAction, ChatMessage } from '../../core/api/models';

const CHAT_KEY = 'marketino.mobile.chat';
const MAX_STORED_MESSAGES = 100;

@Component({
  selector: 'app-mobile-chat',
  standalone: true,
  imports: [],
  templateUrl: './mobile-chat.component.html',
  styleUrl: './mobile-chat.component.scss',
})
export class MobileChatComponent {
  private readonly assistant = inject(AssistantService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);

  private readonly scroller = viewChild.required<ElementRef<HTMLElement>>('scroller');

  readonly messages = signal<ChatMessage[]>(this.readStoredMessages());
  readonly draft = signal('');
  readonly sending = signal(false);
  readonly error = signal('');

  private controller: AbortController | null = null;

  constructor() {
    effect(() => {
      this.messages();
      queueMicrotask(() => {
        const el = this.scroller().nativeElement;
        el.scrollTop = el.scrollHeight;
      });
    });
    effect(() => this.persist(this.messages()));
    this.destroyRef.onDestroy(() => this.abortStream());
  }

  back(): void {
    this.router.navigate(['/m']);
  }

  updateDraft(value: string): void {
    this.draft.set(value);
  }

  clear(): void {
    this.abortStream();
    this.messages.set([]);
    this.error.set('');
  }

  send(): void {
    const text = this.draft().trim();
    if (!text || this.sending()) return;

    const conversation: ChatMessage[] = [
      ...this.messages(),
      { role: 'user', content: text },
    ];

    this.messages.set([...conversation, { role: 'assistant', content: '' }]);
    this.draft.set('');
    this.sending.set(true);
    this.error.set('');

    this.controller = new AbortController();
    let received = '';

    this.assistant
      .streamChat(
        conversation,
        {
          onDelta: (delta) => {
            received += delta;
            this.patchLast({ content: received });
          },
          onAction: (action) => this.appendAction(action),
        },
        this.controller.signal,
      )
      .catch((err: unknown) => {
        if (received === '' && this.messages().at(-1)?.content === '') {
          this.messages.update((list) => list.slice(0, -1));
        }
        const aborted = err instanceof DOMException && err.name === 'AbortError';
        this.error.set(!aborted && err instanceof Error ? err.message : '');
      })
      .finally(() => {
        this.sending.set(false);
        this.controller = null;
      });
  }

  stop(): void {
    this.abortStream();
    const last = this.messages().at(-1);
    if (last && last.role === 'assistant' && last.content.trim() === '' && !last.actions?.length) {
      this.messages.update((list) => list.slice(0, -1));
    }
  }

  private abortStream(): void {
    this.controller?.abort();
    this.controller = null;
  }

  private patchLast(patch: Partial<ChatMessage>): void {
    this.messages.update((list) => {
      if (list.length === 0) return list;
      return [...list.slice(0, -1), { ...list[list.length - 1], ...patch }];
    });
  }

  private appendAction(action: ChatAction): void {
    const last = this.messages().at(-1);
    this.patchLast({ actions: [...(last?.actions ?? []), action] });
  }

  private readStoredMessages(): ChatMessage[] {
    try {
      const raw = localStorage.getItem(CHAT_KEY);
      if (!raw) return [];
      const parsed = JSON.parse(raw) as ChatMessage[];
      return Array.isArray(parsed) ? parsed.slice(-MAX_STORED_MESSAGES) : [];
    } catch {
      return [];
    }
  }

  private persist(messages: ChatMessage[]): void {
    localStorage.setItem(CHAT_KEY, JSON.stringify(messages.slice(-MAX_STORED_MESSAGES)));
  }
}
