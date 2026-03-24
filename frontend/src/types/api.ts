export type SSEStatus = 'connecting' | 'open' | 'closed';

export type WSStatus = 'connecting' | 'open' | 'closed';

export interface SSEOptions {
  withCredentials?: boolean;
  retryInterval?: number;
  maxRetries?: number;
  retryOnClose?: boolean;
  headers?: Record<string, string>;
  onOpen?: () => void;
  onMessage?: (event: MessageEvent) => void;
  onError?: (error: Event) => void;
  onClose?: () => void;
}

export interface SSEEvent<T = unknown> {
  id?: string;
  event?: string;
  data: T;
  retry?: number;
}

export interface WSOptions {
  protocols?: string | string[];
  retryInterval?: number;
  maxRetries?: number;
  retryOnClose?: boolean;
  heartbeatInterval?: number;
  heartbeatMessage?: string | Record<string, unknown>;
  queueMessages?: boolean;
  maxQueueSize?: number;
  onOpen?: (event: Event) => void;
  onMessage?: (event: MessageEvent) => void;
  onError?: (event: Event) => void;
  onClose?: (event: CloseEvent) => void;
  onReconnecting?: (attempt: number) => void;
}

export interface WSMessage<T = unknown> {
  type: string;
  payload: T;
  timestamp?: number;
}

export interface RequestOptions extends RequestInit {
  timeout?: number;
  retryCount?: number;
  retryDelay?: number;
  baseURL?: string;
  params?: Record<string, string | number | boolean>;
}

export interface ApiResponse<T = unknown> {
  data: T;
  status: number;
  statusText: string;
  headers: Headers;
}

export interface ApiError extends Error {
  status?: number;
  statusText?: string;
  response?: ApiResponse;
}
