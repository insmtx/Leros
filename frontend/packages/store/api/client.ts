import type { HttpClient } from "@leros/ui/lib/request";
import { createHttpClient } from "@leros/ui/lib/request";
import { API_BASE_URL } from "./config";

export const apiClient: HttpClient = createHttpClient(API_BASE_URL);
