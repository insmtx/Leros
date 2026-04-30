import { electronAPI } from "@electron-toolkit/preload";

declare global {
	interface Window {
		electron: typeof electronAPI;
	}
}
