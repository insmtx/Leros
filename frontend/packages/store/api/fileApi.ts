import { authenticatedFetch } from "../utils/authStorage";
import { API_BASE_URL } from "./config";

export function getFileDownloadUrl(publicId: string): string {
	return `${API_BASE_URL}/files/${encodeURIComponent(publicId)}/download`;
}

export async function fetchFileDownload(
	publicId: string,
	options?: { signal?: AbortSignal },
): Promise<Response> {
	const response = await authenticatedFetch(getFileDownloadUrl(publicId), {
		method: "GET",
		signal: options?.signal,
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
