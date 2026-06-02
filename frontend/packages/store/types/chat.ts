export type MessageRole = "user" | "assistant" | "system" | "tool";

export type ToolCallStatus = "pending" | "running" | "success" | "error";

export type TodoStatus = "pending" | "in_progress" | "completed" | "cancelled";

export type ToolCall = {
	id: string;
	name: string;
	arguments: Record<string, unknown>;
	status: ToolCallStatus;
	result?: unknown;
	duration?: number;
};

export type RuntimeTodoItem = {
	id: string;
	title: string;
	status: TodoStatus;
	priority?: string;
};

export type MessageArtifact = {
	id: string;
	name: string;
	title: string;
	description?: string;
	type: "document" | "spreadsheet" | "image";
	artifactType: string;
	mimeType?: string;
	size: string;
	updatedAt: string;
	downloadUrl: string;
	sha256?: string;
};

export type MessageMetadata = {
	model?: string;
	tokens?: number;
	latency?: number;
};

export type Message = {
	id: string;
	conversationId: string;
	role: MessageRole;
	content: string;
	timestamp: number;
	sequence?: number;
	toolCalls?: ToolCall[];
	todos?: RuntimeTodoItem[];
	artifacts?: MessageArtifact[];
	thinking?: string;
	metadata?: MessageMetadata;
};

export type Attachment = {
	id: string;
	type: "image" | "file";
	name: string;
	size: number;
	url?: string;
	file?: File;
};

export type ModelOption = {
	id: string;
	label: string;
	provider: string;
};
