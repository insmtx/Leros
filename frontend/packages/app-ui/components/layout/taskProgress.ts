import type { Message, RuntimeTodoItem } from "@leros/store/types/chat";

export function getLatestAssistantTodos(
	messagesMap: Record<string, Message>,
	messageIds: string[],
	sessionId: string | null | undefined,
	streamingMessageId: string | null,
): RuntimeTodoItem[] | undefined {
	if (!sessionId) return undefined;

	const sessionMessages = messageIds
		.map((id) => messagesMap[id])
		.filter((message): message is Message => message?.conversationId === sessionId);

	if (streamingMessageId) {
		const streamingMessage = messagesMap[streamingMessageId];
		if (streamingMessage?.role === "assistant") {
			// Extend here when GetSessionMessages adds a top-level todos field.
			return streamingMessage.todos;
		}
	}

	for (let index = sessionMessages.length - 1; index >= 0; index -= 1) {
		const message = sessionMessages[index];
		if (message?.role === "assistant") {
			return message.todos;
		}
	}

	return undefined;
}
