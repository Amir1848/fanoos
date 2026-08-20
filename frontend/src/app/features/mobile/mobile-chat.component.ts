import { Component, DestroyRef, ElementRef, effect, inject, signal, viewChild } from '@angular/core';
import { Router } from '@angular/router';
import { AssistantService } from '../../core/api/assistant.service';
import { ChatAction, ChatMessage } from '../../core/api/models';
import { audioBlobToWav } from '../../core/audio/wav';

const CHAT_KEY = 'fanoos.mobile.chat';
const MAX_STORED_MESSAGES = 100;
const MAX_RECORDING_MS = 60_000;

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
  readonly recording = signal(false);
  readonly transcribing = signal(false);

  private controller: AbortController | null = null;
  private recorder: MediaRecorder | null = null;
  private recordingStream: MediaStream | null = null;
  private recordTimer: number | null = null;

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
    this.destroyRef.onDestroy(() => this.cleanupRecording());
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

  async toggleMic(): Promise<void> {
    if (this.recording()) {
      this.stopRecording();
      return;
    }
    if (this.sending() || this.transcribing()) return;

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const recorder = new MediaRecorder(stream);
      const chunks: Blob[] = [];

      recorder.ondataavailable = (event: BlobEvent) => {
        if (event.data.size > 0) chunks.push(event.data);
      };
      recorder.onstop = () => this.finishRecording(chunks, stream);

      this.recordingStream = stream;
      this.recorder = recorder;
      this.error.set('');
      recorder.start();
      this.recording.set(true);
      this.recordTimer = window.setTimeout(() => this.stopRecording(), MAX_RECORDING_MS);
    } catch (err) {
      this.recordingStream?.getTracks().forEach((track) => track.stop());
      this.recordingStream = null;
      this.error.set(err instanceof DOMException && err.name === 'NotAllowedError'
        ? 'دسترسی به میکروفون داده نشد'
        : 'ضبط صدا ممکن نیست');
    }
  }

  private stopRecording(): void {
    if (this.recordTimer !== null) {
      clearTimeout(this.recordTimer);
      this.recordTimer = null;
    }
    this.recording.set(false);
    this.recorder?.stop();
    this.recorder = null;
  }

  private cleanupRecording(): void {
    if (this.recordTimer !== null) {
      clearTimeout(this.recordTimer);
      this.recordTimer = null;
    }
    if (this.recorder) {
      this.recorder.onstop = null;
      this.recorder.stop();
      this.recorder = null;
    }
    this.recordingStream?.getTracks().forEach((track) => track.stop());
    this.recordingStream = null;
  }

  private async finishRecording(chunks: Blob[], stream: MediaStream): Promise<void> {
    stream.getTracks().forEach((track) => track.stop());
    this.recordingStream = null;
    if (chunks.length === 0) return;

    this.transcribing.set(true);
    try {
      const wav = await audioBlobToWav(new Blob(chunks));
      const text = await this.assistant.transcribe(wav);
      if (text) this.draft.set(text);
    } catch (err) {
      this.error.set(err instanceof Error && err.message ? err.message : 'خطا در تبدیل صدا به متن');
    } finally {
      this.transcribing.set(false);
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
