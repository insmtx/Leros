import { readStoredJwtToken } from "../utils/authStorage";
import { apiClient } from "./client";
import { API_BASE_URL } from "./config";
import type {
	BackendDataResponse,
	BackendProjectFileNode,
	BackendProjectFileUploadResult,
} from "./types";

export type GetProjectFilesParams = {
	projectId: string;
	path?: string;
	depth?: number;
};

export type UploadProjectFileParams = {
	projectId: string;
	file: File;
};

export function getProjectFileDownloadUrl(projectId: string, filepath: string): string {
	const normalizedPath = filepath.replace(/^\/+/, "");
	return `${API_BASE_URL}/projects/${encodeURIComponent(projectId)}/files/${normalizedPath
		.split("/")
		.map((segment) => encodeURIComponent(segment))
		.join("/")}`;
}

export async function fetchProjectFileDownload(
	projectId: string,
	filepath: string,
	options?: { signal?: AbortSignal },
): Promise<Response> {
	const token = readStoredJwtToken();
	const response = await fetch(getProjectFileDownloadUrl(projectId, filepath), {
		method: "GET",
		signal: options?.signal,
		headers: token ? { Authorization: `Bearer ${token}` } : undefined,
	});
	if (!response.ok) {
		throw new Error(`HTTP ${response.status}`);
	}
	return response;
}

export const projectFileApi = {
	list: ({ projectId, path, depth = 2 }: GetProjectFilesParams) =>
		apiClient.get<BackendDataResponse<BackendProjectFileNode[]>>(
			`/projects/${encodeURIComponent(projectId)}/files`,
			{
				params: {
					...(path ? { path } : {}),
					depth,
				},
			},
		),

	upload: async ({ projectId, file }: UploadProjectFileParams) => {
		const token = readStoredJwtToken();
		const formData = new FormData();
		formData.append("file", file);

		const response = await fetch(
			`${API_BASE_URL}/projects/${encodeURIComponent(projectId)}/files/upload`,
			{
				method: "POST",
				body: formData,
				headers: token ? { Authorization: `Bearer ${token}` } : undefined,
			},
		);

		if (!response.ok) {
			let message = `HTTP ${response.status}`;
			try {
				const payload = (await response.json()) as { message?: string };
				if (typeof payload.message === "string" && payload.message) {
					message = payload.message;
				}
			} catch {
				// 保持默认错误信息即可
			}
			throw new Error(message);
		}

		return (await response.json()) as BackendDataResponse<BackendProjectFileUploadResult | string>;
	},

	getDownloadUrl: getProjectFileDownloadUrl,
	fetchDownload: fetchProjectFileDownload,
};
