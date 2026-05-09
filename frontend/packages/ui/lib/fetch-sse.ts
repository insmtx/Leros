export type FetchSSEStatus = "connecting" | "open" | "closed";

export type FetchSSEOptions = {
	method?: string;
	body?: unknown;
	headers?: Record<string, string>;
	withCredentials?: boolean;
	onOpen?: () => void;
	onMessage?: (event: { type?: string; data: string; id?: string; retry?: number }) => void;
	onError?: (error: Error) => void;
	onClose?: () => void;
};

export class FetchSSEClient {
	private url: string;
	private options: FetchSSEOptions;
	private status: FetchSSEStatus = "closed";
	private abortController: AbortController | null = null;
	private reader: ReadableStreamDefaultReader<Uint8Array> | null = null;

	constructor(url: string, options: FetchSSEOptions = {}) {
		this.url = url;
		this.options = {
			method: "POST",
			...options,
		};
	}

	async connect(): Promise<void> {
		if (this.status === "open" || this.status === "connecting") return;

		this.status = "connecting";
		this.abortController = new AbortController();

		const fetchOptions: RequestInit = {
			method: this.options.method ?? "POST",
			headers: {
				"Content-Type": "application/json",
				Accept: "text/event-stream",
				...this.options.headers,
			},
			body: this.options.body ? JSON.stringify(this.options.body) : undefined,
			signal: this.abortController.signal,
		};

		if (this.options.withCredentials) {
			fetchOptions.credentials = "include";
		}

		try {
			const response = await fetch(this.url, fetchOptions);

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}: ${response.statusText}`);
			}

			if (!response.body) {
				throw new Error("Response body is null");
			}

			this.status = "open";
			this.options.onOpen?.();

			this.reader = response.body.getReader();
			const decoder = new TextDecoder();
			let buffer = "";

			while (true) {
				const { done, value } = await this.reader.read();
				if (done) break;

				buffer += decoder.decode(value, { stream: true });

				const lines = buffer.split("\n");
				buffer = lines.pop() ?? "";

				let currentEvent: { type?: string; data: string; id?: string; retry?: number } = {
					data: "",
				};

				for (const line of lines) {
					if (line.startsWith("event:")) {
						currentEvent.type = line.slice(6).trim();
					} else if (line.startsWith("data:")) {
						currentEvent.data += line.slice(5).trim();
					} else if (line.startsWith("id:")) {
						currentEvent.id = line.slice(3).trim();
					} else if (line.startsWith("retry:")) {
						currentEvent.retry = Number.parseInt(line.slice(6).trim(), 10);
					} else if (line === "") {
						if (currentEvent.data) {
							this.options.onMessage?.(currentEvent);
						}
						currentEvent = { data: "" };
					}
				}
			}

			this.close();
		} catch (err) {
			if ((err as Error).name === "AbortError") {
				this.close();
				return;
			}
			this.status = "closed";
			this.options.onError?.(err as Error);
			this.options.onClose?.();
		}
	}

	close(): void {
		this.status = "closed";

		if (this.reader) {
			this.reader.cancel().catch(() => { /* ignore cancel errors */ });
			this.reader = null;
		}

		if (this.abortController) {
			this.abortController.abort();
			this.abortController = null;
		}

		this.options.onClose?.();
	}

	getStatus(): FetchSSEStatus {
		return this.status;
	}
}

export function createFetchSSE(url: string, options?: FetchSSEOptions): FetchSSEClient {
	return new FetchSSEClient(url, options);
}
