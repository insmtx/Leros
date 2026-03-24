import type { SSEOptions, SSEStatus } from '@/types/api';

type EventHandler = (event: MessageEvent) => void;

export class SSEClient {
  private url: string;
  private options: SSEOptions;
  private eventSource: EventSource | null = null;
  private status: SSEStatus = 'closed';
  private retryCount = 0;
  private eventHandlers: Map<string, Set<EventHandler>> = new Map();
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

  constructor(url: string, options: SSEOptions = {}) {
    this.url = url;
    this.options = {
      retryInterval: 3000,
      maxRetries: 5,
      retryOnClose: false,
      withCredentials: false,
      ...options,
    };
  }

  connect(): void {
    if (this.eventSource) {
      return;
    }

    this.setStatus('connecting');
    this.createEventSource();
  }

  private createEventSource(): void {
    const urlObj = new URL(this.url, window.location.origin);

    if (this.options.headers) {
      const headers = this.options.headers;
      Object.keys(headers).forEach((key) => {
        urlObj.searchParams.append(key, headers[key]);
      });
    }

    this.eventSource = new EventSource(urlObj.toString(), {
      withCredentials: this.options.withCredentials,
    });

    this.setupEventListeners();
  }

  private setupEventListeners(): void {
    if (!this.eventSource) return;

    this.eventSource.onopen = () => {
      this.setStatus('open');
      this.retryCount = 0;
      this.options.onOpen?.();
    };

    this.eventSource.onerror = (error) => {
      this.setStatus('closed');
      this.options.onError?.(error);

      if (this.shouldReconnect()) {
        this.scheduleReconnect();
      }
    };

    this.eventSource.onmessage = (event) => {
      this.options.onMessage?.(event);
      this.dispatchEvent('message', event);
    };
  }

  private shouldReconnect(): boolean {
    const { maxRetries, retryOnClose } = this.options;

    if (!retryOnClose) return false;
    if (maxRetries !== undefined && this.retryCount >= maxRetries) {
      return false;
    }

    return true;
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }

    this.retryCount++;
    const delay = this.options.retryInterval ?? 3000;

    this.reconnectTimeout = setTimeout(() => {
      this.close(false);
      this.connect();
    }, delay);
  }

  private setStatus(status: SSEStatus): void {
    this.status = status;
  }

  private dispatchEvent(type: string, event: MessageEvent): void {
    const handlers = this.eventHandlers.get(type);
    if (handlers) {
      handlers.forEach((handler) => {
        handler(event);
      });
    }
  }

  on(event: string, handler: EventHandler): () => void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
      this.eventSource?.addEventListener(event, (e) => {
        this.dispatchEvent(event, e as MessageEvent);
      });
    }

    this.eventHandlers.get(event)?.add(handler);

    return () => this.off(event, handler);
  }

  off(event: string, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.delete(handler);
    }
  }

  getStatus(): SSEStatus {
    return this.status;
  }

  close(callOnClose = true): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }

    this.setStatus('closed');
    this.retryCount = 0;

    if (callOnClose) {
      this.options.onClose?.();
    }
  }

  reconnect(): void {
    this.close(false);
    this.connect();
  }
}

export function createSSE(url: string, options?: SSEOptions): SSEClient {
  return new SSEClient(url, options);
}
