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

type BackendUploadFilePayload = {
	public_id: string;
	file_upload_id?: string;
	filename?: string;
	original_name?: string;
	mime_type?: string;
	file_size?: number;
	sha256?: string;
	storage_path?: string;
	url?: string;
};

type AddProjectFileParams = {
	projectId: string;
	publicId: string;
};

async function parseErrorMessage(response: Response): Promise<string> {
	let message = `HTTP ${response.status}`;
	try {
		const payload = (await response.json()) as { message?: string };
		if (typeof payload.message === "string" && payload.message) {
			message = payload.message;
		}
	} catch {
		// 保持默认错误信息即可
	}
	return message;
}

function assertBackendSuccess<T>(
	response: BackendDataResponse<T>,
	fallbackMessage: string,
): BackendDataResponse<T> {
	if (response.code !== 0) {
		throw new Error(response.message || fallbackMessage);
	}
	return response;
}

async function uploadFile(
	file: File,
	token?: string | null,
): Promise<BackendDataResponse<BackendUploadFilePayload>> {
	const formData = new FormData();
	formData.append("file", file);
	formData.append("purpose", "project");

	const response = await fetch(`${API_BASE_URL}/files/upload`, {
		method: "POST",
		body: formData,
		headers: token ? { Authorization: `Bearer ${token}` } : undefined,
	});

	if (!response.ok) {
		throw new Error(await parseErrorMessage(response));
	}

	return (await response.json()) as BackendDataResponse<BackendUploadFilePayload>;
}

async function addFileToProject(
	{ projectId, publicId }: AddProjectFileParams,
	token?: string | null,
): Promise<BackendDataResponse<null>> {
	const response = await fetch(
		`${API_BASE_URL}/projects/${encodeURIComponent(projectId)}/AddFile`,
		{
			method: "POST",
			headers: {
				"Content-Type": "application/json",
				...(token ? { Authorization: `Bearer ${token}` } : {}),
			},
			body: JSON.stringify({ public_id: publicId }),
		},
	);

	if (!response.ok) {
		throw new Error(await parseErrorMessage(response));
	}

	return (await response.json()) as BackendDataResponse<null>;
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
		const uploadResponse = assertBackendSuccess(await uploadFile(file, token), "文件上传失败");
		const uploaded = uploadResponse.data;
		if (!uploaded?.public_id) {
			throw new Error("上传接口未返回 public_id");
		}

		const addResponse = assertBackendSuccess(
			await addFileToProject({ projectId, publicId: uploaded.public_id }, token),
			"添加文件到项目失败",
		);

		return {
			code: addResponse.code,
			message: addResponse.message,
			data: {
				path: uploaded.storage_path || uploaded.url || uploaded.public_id,
				filename: uploaded.original_name || uploaded.filename || file.name,
				size: uploaded.file_size ?? file.size,
				public_id: uploaded.public_id,
				file_upload_id: uploaded.file_upload_id,
				original_name: uploaded.original_name,
				mime_type: uploaded.mime_type || file.type,
				file_size: uploaded.file_size ?? file.size,
				sha256: uploaded.sha256,
				storage_path: uploaded.storage_path,
				url: uploaded.url,
			} satisfies BackendProjectFileUploadResult,
		} as BackendDataResponse<BackendProjectFileUploadResult>;
	},
};
