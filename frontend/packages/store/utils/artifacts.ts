import type { ProjectArtifact } from "../slices/layoutSlice";
import type { Message, MessageArtifact } from "../types/chat";

/** Converts a message-scoped artifact into the project artifact shape used by UI panels. */
export function messageArtifactToProjectArtifact(artifact: MessageArtifact): ProjectArtifact {
	return {
		id: artifact.id,
		name: artifact.name,
		title: artifact.title,
		description: artifact.description,
		type: artifact.type,
		artifactType: artifact.artifactType,
		mimeType: artifact.mimeType,
		size: artifact.size,
		updatedAt: artifact.updatedAt,
		downloadUrl: artifact.downloadUrl,
		sha256: artifact.sha256,
	};
}

/**
 * Merges session message artifacts with task API artifacts.
 * Task API records enrich message artifacts when both exist for the same id.
 */
export function mergeProjectArtifacts(
	taskArtifacts: ProjectArtifact[],
	sessionArtifacts: ProjectArtifact[],
): ProjectArtifact[] {
	const merged = new Map<string, ProjectArtifact>();
	for (const artifact of sessionArtifacts) {
		merged.set(artifact.id, artifact);
	}
	for (const artifact of taskArtifacts) {
		const existing = merged.get(artifact.id);
		merged.set(artifact.id, existing ? { ...existing, ...artifact } : artifact);
	}
	return [...merged.values()];
}

/** Collects declared artifacts from assistant messages in one session. */
export function collectSessionArtifacts(
	messagesMap: Record<string, Message>,
	messageIds: string[],
	sessionId: string | null | undefined,
): ProjectArtifact[] {
	if (!sessionId) return [];

	const merged = new Map<string, ProjectArtifact>();
	for (const id of messageIds) {
		const message = messagesMap[id];
		if (
			!message ||
			message.conversationId !== sessionId ||
			message.role !== "assistant" ||
			!message.artifacts?.length
		) {
			continue;
		}
		for (const artifact of message.artifacts) {
			merged.set(artifact.id, messageArtifactToProjectArtifact(artifact));
		}
	}
	return [...merged.values()];
}
