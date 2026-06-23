import type { BackendDataResponse } from "./types";
import { apiClient } from "./client";
import {
	dispatchClientUpgradeRequired,
	getClientVersionReport,
	type ClientUpdatePolicy,
	type ClientVersionReportParams,
} from "./clientUpdatePolicy";

export const CLIENT_UPDATE_ENDPOINTS = {
	report: "/ClientVersionReport",
} as const;

export const clientUpdateApi = {
	reportVersion: async (params: ClientVersionReportParams = getClientVersionReport()) => {
		const response = await apiClient.post<BackendDataResponse<ClientUpdatePolicy>>(
			CLIENT_UPDATE_ENDPOINTS.report,
			params,
		);

		if (response.data.data?.force_update) {
			dispatchClientUpgradeRequired(response.data.data);
		}

		return response;
	},
};
