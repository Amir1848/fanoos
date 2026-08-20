import { Injectable, inject } from '@angular/core';
import { AuthService } from '../auth/auth.service';
import { ChatAction, ChatMessage } from './models';

export interface ChatStreamCallbacks {
  onDelta: (delta: string) => void;
  onAction?: (action: ChatAction) => void;
}

@Injectable({ providedIn: 'root' })
export class AssistantService {
  private readonly auth = inject(AuthService);

  async transcribe(audio: Blob): Promise<string> {
    const form = new FormData();
    form.append('audio', audio, 'audio.wav');

    const headers: Record<string, string> = {};
    if (this.auth.token) {
      headers['Authorization'] = `Bearer ${this.auth.token}`;
    }

    const res = await fetch('/api/assistant/transcribe', {
      method: 'POST',
      headers,
      body: form,
    });

    if (!res.ok) {
      if (res.status === 401) {
        this.auth.logout();
      }
      throw new Error(await readErrorMessage(res, 'خطا در تبدیل صدا به متن'));
    }

    const body = (await res.json()) as { text?: string };
    return (body.text ?? '').trim();
  }

  async streamChat(
    messages: ChatMessage[],
    callbacks: ChatStreamCallbacks,
    signal?: AbortSignal,
  ): Promise<void> {
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    if (this.auth.token) {
      headers['Authorization'] = `Bearer ${this.auth.token}`;
    }

    const res = await fetch('/api/assistant/chat', {
      method: 'POST',
      headers,
      body: JSON.stringify({ messages }),
      signal,
    });

    if (!res.ok) {
      if (res.status === 401) {
        this.auth.logout();
      }
      throw new Error(await readErrorMessage(res, 'خطا در برقراری ارتباط با دستیار'));
    }

    if (!res.body) {
      throw new Error('پاسخی از سرور دریافت نشد');
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const events = buffer.split('\n\n');
      buffer = events.pop() ?? '';

      for (const event of events) {
        const line = event
          .split('\n')
          .find((l) => l.startsWith('data:'));
        if (!line) continue;

        const data = line.slice('data:'.length).trim();
        if (data === '[DONE]') return;

        try {
          const parsed = JSON.parse(data) as { delta?: string; error?: string; action?: ChatAction };
          if (parsed.error) throw new Error(parsed.error);
          if (parsed.action) callbacks.onAction?.(parsed.action);
          if (parsed.delta) callbacks.onDelta(parsed.delta);
        } catch (err) {
          if (err instanceof SyntaxError) continue;
          throw err;
        }
      }
    }
  }
}

async function readErrorMessage(res: Response, fallback: string): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string };
    return body.error || fallback;
  } catch {
    return fallback;
  }
}
