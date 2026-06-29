import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { AIMessageBubble } from "./AIMessageBubble";

vi.mock("@leros/store", () => ({
	formatArtifactTime: () => "",
	formatTime: () => "10:00",
	getAssistantMessageFooterSegments: () => [],
	messageArtifactToProjectArtifact: vi.fn(),
	sortProjectArtifactsByNewestFirst: (artifacts: unknown[]) => artifacts,
	useChatStore: (selector: (state: Record<string, unknown>) => unknown) =>
		selector({
			resendMessage: vi.fn(),
		}),
}));

vi.mock("../common/MarkdownRenderer", () => ({
	MarkdownRenderer: ({ content }: { content: string }) => <div>{content}</div>,
}));

vi.mock("../layout/ArtifactPreviewDialog", () => ({
	ArtifactPreviewDialog: () => null,
}));

vi.mock("../layout/project-file-type-icon", () => ({
	ProjectFileTypeIcon: () => null,
}));

vi.mock("./AssistantChatAvatar", () => ({
	AssistantChatAvatar: () => <div>avatar</div>,
}));

describe("AIMessageBubble", () => {
	it("执行过程默认收起，且流式状态变化不会覆盖用户手动展开", async () => {
		const user = userEvent.setup();
		const message = {
			id: "message-1",
			conversationId: "conversation-1",
			role: "assistant" as const,
			content: "",
			timestamp: Date.now(),
			processSteps: [
				{
					id: "step-1",
					type: "thinking" as const,
					content: "正在分析问题",
				},
			],
			toolCalls: [],
		};

		const { rerender } = render(<AIMessageBubble message={message} isStreaming={true} />);

		expect(screen.getByRole("button", { name: /执行过程/i })).toBeInTheDocument();
		expect(screen.queryByText("正在分析问题", { selector: "div" })).not.toBeInTheDocument();

		await user.click(screen.getByRole("button", { name: /执行过程/i }));

		expect(screen.getByText("正在分析问题", { selector: "div" })).toBeInTheDocument();

		rerender(<AIMessageBubble message={message} isStreaming={false} />);

		expect(screen.getByText("正在分析问题", { selector: "div" })).toBeInTheDocument();
	});

	it("执行过程收起时展示最新的过程摘要", () => {
		const message = {
			id: "message-2",
			conversationId: "conversation-1",
			role: "assistant" as const,
			content: "",
			timestamp: Date.now(),
			processSteps: [
				{ id: "tool-call-1", type: "tool_call" as const, toolCallId: "tool-call-1" },
				{ id: "thinking-1", type: "thinking" as const, content: "正在整理文档结构" },
			],
			toolCalls: [
				{
					id: "tool-call-1",
					name: "skill",
					arguments: {},
					status: "running" as const,
				},
			],
		};

		const { rerender } = render(<AIMessageBubble message={message} isStreaming={true} />);

		expect(screen.getByText("正在整理文档结构")).toBeInTheDocument();
		expect(screen.queryByText("调用：skill")).not.toBeInTheDocument();

		rerender(
			<AIMessageBubble
				message={{
					...message,
					processSteps: [
						...message.processSteps.slice(0, -1),
						{ id: "thinking-1", type: "thinking", content: "正在写入最终文档" },
					],
				}}
				isStreaming={true}
			/>,
		);

		expect(screen.getByText("正在写入最终文档")).toBeInTheDocument();
		expect(screen.queryByText("正在整理文档结构")).not.toBeInTheDocument();
	});
});
