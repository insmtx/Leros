import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { SSEOptions, SSEStatus } from "../lib/sse";
import { SSEClient } from "../lib/sse";

interface UseSSEReturn<T> {
	data: T | null;
	status: SSEStatus;
	error: Event | null;
	reconnect: () => void;
}

export function useSSE<T = unknown>(url: string | null, options: SSEOptions = {}): UseSSEReturn<T> {
	const [data, setData] = useState<T | null>(null);
	const [status, setStatus] = useState<SSEStatus>("closed");
	const [error, setError] = useState<Event | null>(null);

	const clientRef = useRef<SSEClient | null>(null);
	const optionsRef = useRef(options);
	optionsRef.current = options;

	const stableOptions = useMemo(
		() => ({
			onOpen: optionsRef.current.onOpen,
			onMessage: optionsRef.current.onMessage,
			onError: optionsRef.current.onError,
			onClose: optionsRef.current.onClose,
			withCredentials: optionsRef.current.withCredentials,
			retryInterval: optionsRef.current.retryInterval,
			maxRetries: optionsRef.current.maxRetries,
			retryOnClose: optionsRef.current.retryOnClose,
			headers: optionsRef.current.headers,
		}),
		[],
	);

	useEffect(() => {
		if (!url) {
			setData(null);
			setStatus("closed");
			setError(null);
			return;
		}

		const client = new SSEClient(url, {
			...stableOptions,
			onOpen: () => {
				setStatus("open");
				setError(null);
				stableOptions.onOpen?.();
			},
			onMessage: (event) => {
				try {
					const parsed = JSON.parse(event.data) as T;
					setData(parsed);
				} catch {
					setData(event.data as unknown as T);
				}
				stableOptions.onMessage?.(event);
			},
			onError: (err) => {
				setStatus("closed");
				setError(err);
				stableOptions.onError?.(err);
			},
			onClose: () => {
				setStatus("closed");
				stableOptions.onClose?.();
			},
		});

		clientRef.current = client;
		client.connect();

		return () => {
			client.close();
			clientRef.current = null;
		};
	}, [url, stableOptions]);

	const reconnect = useCallback(() => {
		clientRef.current?.reconnect();
	}, []);

	return { data, status, error, reconnect };
}
