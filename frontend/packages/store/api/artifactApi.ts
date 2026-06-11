import type { ApiError } from "@leros/ui/lib/request";
import { apiClient } from "./client";
import { fetchFileDownload } from "./fileApi";
import type { BackendArtifact, BackendArtifactDetail, BackendDataResponse } from "./types";

function isNotFoundError(error: unknown): boolean {
	return (
		typeof error === "object" &&
		error !== null &&
		"status" in error &&
		(error as ApiError).status === 404
	);
}

/** Lists task artifacts; prefers deployed RPC route, falls back to REST GET for local dev. */
async function listTaskArtifacts(taskId: string) {
	const normalizedTaskId = taskId.trim();
	if (!normalizedTaskId) {
		throw new Error("task_id is required");
	}

	try {
		return await apiClient.post<BackendDataResponse<BackendArtifact[]>>("/ListTaskArtifacts", {
			task_id: normalizedTaskId,
		});
	} catch (error) {
		if (!isNotFoundError(error)) throw error;
		return apiClient.get<BackendDataResponse<BackendArtifact[]>>(
			`/tasks/${encodeURIComponent(normalizedTaskId)}/artifacts`,
		);
	}
}

const publishFileIdCache = new Map<string, string>();

function readPublishFileId(detail: BackendArtifactDetail): string {
	const publishFileId = detail.publish_file_id?.trim() ?? detail["publish-file_id"]?.trim() ?? "";
	return publishFileId;
}

async function resolveArtifactPublishFileId(
	artifactId: string,
	options?: { signal?: AbortSignal },
): Promise<string> {
	const normalizedArtifactId = artifactId.trim();
	if (!normalizedArtifactId) {
		throw new Error("artifact_id is required");
	}

	const cached = publishFileIdCache.get(normalizedArtifactId);
	if (cached) return cached;

	const response = await apiClient.post<BackendDataResponse<BackendArtifactDetail>>(
		"/GetArtifact",
		{ artifact_id: normalizedArtifactId },
		{ signal: options?.signal },
	);
	const publishFileId = readPublishFileId(response.data.data ?? {});
	if (!publishFileId) {
		throw new Error("GetArtifact 未返回 publish_file_id");
	}

	publishFileIdCache.set(normalizedArtifactId, publishFileId);
	return publishFileId;
}

export async function fetchArtifactDownload(
	artifactId: string,
	options?: { signal?: AbortSignal },
): Promise<Response> {
	const publishFileId = await resolveArtifactPublishFileId(artifactId, options);
	return fetchFileDownload(publishFileId, options);
}

export const artifactApi = {
	fetchDownload: fetchArtifactDownload,
	listTaskArtifacts,
};
