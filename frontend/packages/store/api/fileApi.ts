import { readStoredJwtToken } from "../utils/authStorage";
import { API_BASE_URL } from "./config";

export function getFileDownloadUrl(publicId: string): string {
	return `${API_BASE_URL}/files/${encodeURIComponent(publicId)}/download`;
}

export async function fetchFileDownload(
	publicId: string,
	options?: { signal?: AbortSignal },
): Promise<Response> {
	const token = readStoredJwtToken();
	const response = await fetch(getFileDownloadUrl(publicId), {
		method: "GET",
		signal: options?.signal,
		headers: token ? { Authorization: `Bearer ${token}` } : undefined,
	});
	if (!response.ok) {
		throw new Error(`HTTP ${response.status}`);
	}
	return response;
}

export const fileApi = {
	getDownloadUrl: getFileDownloadUrl,
	fetchDownload: fetchFileDownload,
};
