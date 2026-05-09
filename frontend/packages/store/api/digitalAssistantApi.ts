import { apiClient } from "./client";
import type {
	BackendDataResponse,
	BackendDigitalAssistant,
	BackendPaginatedResponse,
} from "./types";

export type CreateDAParams = {
	code: string;
	name: string;
	org_id: number;
	owner_id: number;
	description?: string;
	avatar?: string;
	config?: {
		llm_config?: { type?: string };
		skills?: { skill_code?: string; version?: string }[];
		channels?: { type?: string }[];
		knowledge?: { type?: string; repo?: string; dataset_id?: string }[];
		memory_config?: { type?: string };
		policies_config?: { type?: string };
		runtime_config?: { type?: string };
	};
};

export type UpdateDAParams = {
	id: number;
	name?: string;
	description?: string;
	avatar?: string;
	config?: {
		llm_config?: { type?: string };
		skills?: { skill_code?: string; version?: string }[];
		channels?: { type?: string }[];
		knowledge?: { type?: string; repo?: string; dataset_id?: string }[];
		memory_config?: { type?: string };
		policies_config?: { type?: string };
		runtime_config?: { type?: string };
	};
};

export type UpdateDAConfigParams = {
	id: number;
	config: {
		llm_config?: { type?: string };
		skills?: { skill_code?: string; version?: string }[];
		channels?: { type?: string }[];
		knowledge?: { type?: string; repo?: string; dataset_id?: string }[];
		memory_config?: { type?: string };
		policies_config?: { type?: string };
		runtime_config?: { type?: string };
	};
};

export type UpdateDAStatusParams = {
	id: number;
	status: string;
};

export type ListDAParams = {
	page?: number;
	per_page?: number;
	org_id?: number;
	owner_id?: number;
	status?: string;
	keyword?: string;
};

export type GetDAParams = {
	id?: number;
	code?: string;
};

const DA_ENDPOINTS = {
	create: "/CreateDigitalAssistant",
	list: "/ListDigitalAssistant",
	get: "/GetDigitalAssistant",
	update: "/UpdateDigitalAssistant",
	updateConfig: "/UpdateDigitalAssistantConfig",
	updateStatus: "/UpdateDigitalAssistantStatus",
	delete: "/DeleteDigitalAssistant",
};

export const digitalAssistantApi = {
	create: (params: CreateDAParams) =>
		apiClient.post<BackendDataResponse<BackendDigitalAssistant>>(DA_ENDPOINTS.create, params),

	list: (params: ListDAParams) =>
		apiClient.post<BackendPaginatedResponse<BackendDigitalAssistant>>(DA_ENDPOINTS.list, params),

	get: (params: GetDAParams) =>
		apiClient.post<BackendDataResponse<BackendDigitalAssistant>>(DA_ENDPOINTS.get, params),

	update: (params: UpdateDAParams) =>
		apiClient.post<BackendDataResponse<BackendDigitalAssistant>>(DA_ENDPOINTS.update, params),

	updateConfig: (params: UpdateDAConfigParams) =>
		apiClient.post<BackendDataResponse<BackendDigitalAssistant>>(DA_ENDPOINTS.updateConfig, params),

	updateStatus: (params: UpdateDAStatusParams) =>
		apiClient.post<BackendDataResponse<null>>(DA_ENDPOINTS.updateStatus, params),

	delete: (id: number) => apiClient.post<BackendDataResponse<null>>(DA_ENDPOINTS.delete, { id }),
};
