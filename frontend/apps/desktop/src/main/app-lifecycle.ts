import { BrowserWindow } from "electron";

let appQuitting = false;
let appQuitPrepared = false;
let appQuitPreparation: Promise<void> | null = null;

const leaveFullScreenTimeoutMs = 1200;

export function isAppQuitting(): boolean {
	return appQuitting;
}

export function markAppQuitting(): void {
	appQuitting = true;
}

export function isAppQuitPrepared(): boolean {
	return appQuitPrepared;
}

export async function prepareForAppQuit(): Promise<void> {
	if (appQuitPrepared) {
		return;
	}

	if (!appQuitPreparation) {
		appQuitPreparation = Promise.all(BrowserWindow.getAllWindows().map(prepareWindowForQuit)).then(
			() => {
				appQuitPrepared = true;
			},
		);
	}

	await appQuitPreparation;
}

export async function prepareWindowForHide(window: BrowserWindow): Promise<void> {
	if (window.isDestroyed()) {
		return;
	}

	if (window.isFullScreen()) {
		await leaveFullScreen(window);
	}
}

async function prepareWindowForQuit(window: BrowserWindow): Promise<void> {
	if (window.isDestroyed()) {
		return;
	}

	if (window.isFullScreen()) {
		await leaveFullScreen(window);
	}

	if (window.isDestroyed()) {
		return;
	}

	if (window.isMinimized()) {
		window.restore();
	}

	if (!window.isVisible()) {
		window.showInactive();
	}
}

function leaveFullScreen(window: BrowserWindow): Promise<void> {
	return new Promise((resolve) => {
		let timeout: NodeJS.Timeout | null = null;

		const cleanup = () => {
			window.off("leave-full-screen", cleanup);
			if (timeout) {
				clearTimeout(timeout);
				timeout = null;
			}
			resolve();
		};

		window.once("leave-full-screen", cleanup);
		timeout = setTimeout(cleanup, leaveFullScreenTimeoutMs);
		window.setFullScreen(false);
	});
}
