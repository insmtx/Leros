import { apiClient } from "./client";
import type { BackendDataResponse } from "./types";

export interface SkillMarketplaceItem {
  source_type: string;
  skill_id: string;
  name: string;
  description: string;
  version: string;
  author: string;
  category: string;
  tags: string[] | null;
  icon: string;
  installs: number;
}

export interface SearchSkillMarketplaceResponse {
  items: SkillMarketplaceItem[];
  warnings?: Array<{ source_type: string; message: string }>;
}

export interface SearchSkillMarketplaceParams {
  keyword?: string;
  category?: string;
  source_types?: string[];
  limit?: number;
}

export interface InstallSkillParams {
  source: string;
  skill_id: string;
}

export interface InstallSkillResponse {
  status: string;
  message: string;
}

function cleanParams(
  params: SearchSkillMarketplaceParams,
): Record<string, string | number | boolean | string[]> {
  const result: Record<string, string | number | boolean | string[]> = {};
  if (params.keyword) result.keyword = params.keyword;
  if (params.category) result.category = params.category;
  if (params.source_types?.length) result.source_types = params.source_types;
  if (params.limit !== undefined) result.limit = params.limit;
  return result;
}

export const skillMarketplaceApi = {
  search: (params: SearchSkillMarketplaceParams) =>
    apiClient.get<BackendDataResponse<SearchSkillMarketplaceResponse>>(
      "/skill-marketplace/search",
      { params: cleanParams(params) },
    ),

  install: (params: InstallSkillParams) =>
    apiClient.post<BackendDataResponse<InstallSkillResponse>>(
      "/skill-marketplace/install",
      params,
    ),
};
