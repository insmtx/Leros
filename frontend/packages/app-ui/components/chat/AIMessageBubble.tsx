"use client";

import { formatTime, useChatStore } from "@leros/store";
import type { Message } from "@leros/store/types/chat";
import { Avatar, AvatarFallback } from "@leros/ui/components/ui/avatar";
import { Button } from "@leros/ui/components/ui/button";
import { Check, Copy, RefreshCw } from "lucide-react";
import { useState } from "react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ToolCallBlock } from "./ToolCallBlock";

function CopyButton({ text }: { text: string }) {
	const [copied, setCopied] = useState(false);
	const handleCopy = () => {
		navigator.clipboard.writeText(text);
		setCopied(true);
		setTimeout(() => setCopied(false), 1500);
	};
	return (
		<Button
			variant="ghost"
			size="icon-xs"
			className={copied ? "text-green-500" : "text-slate-400 hover:text-slate-600"}
			onClick={handleCopy}
		>
			{copied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
		</Button>
	);
}

export function AIMessageBubble({
	message,
	isStreaming,
}: {
	message: Message;
	isStreaming: boolean;
}) {
	const { resendMessage } = useChatStore((s) => s);
	const content = message.content;
	const hasContent = content.trim().length > 0;
	const hasToolCalls = message.toolCalls && message.toolCalls.length > 0;

	return (
		<div data-slot="ai-message" className="group flex items-start gap-3">
			<Avatar size="sm">
				<AvatarFallback className="bg-blue-500 text-white text-xs">AI</AvatarFallback>
			</Avatar>
			<div className="min-w-0 flex-1">
				<div className="mb-1.5 flex items-center gap-2">
					<span className="text-xs font-medium text-slate-500">AI 助手</span>
					<span className="text-xs text-slate-400">{formatTime(message.timestamp)}</span>
					{isStreaming && <span className="text-xs text-blue-500 animate-pulse">生成中</span>}
				</div>

				{hasToolCalls && message.toolCalls && (
					<div className="mb-3">
						<ToolCallBlock toolCalls={message.toolCalls} />
					</div>
				)}

				{hasContent && (
					<div className="w-fit max-w-[min(780px,92%)] rounded-2xl rounded-tl-md bg-white/90 px-4 py-3 text-sm leading-7 text-slate-700 shadow-sm ring-1 ring-slate-200/50">
						<div className="prose prose-slate prose-sm max-w-none prose-p:my-1.5 prose-pre:my-2 prose-ul:my-1.5 prose-ol:my-1.5">
							<Markdown remarkPlugins={[remarkGfm]}>{content}</Markdown>
						</div>
						{isStreaming && (
							<span className="inline-block w-1.5 h-4 bg-slate-400 animate-pulse ml-0.5 rounded-sm" />
						)}
					</div>
				)}

				{!hasContent && !hasToolCalls && isStreaming && (
					<div className="w-fit rounded-2xl rounded-tl-md bg-white/90 px-4 py-3 shadow-sm ring-1 ring-slate-200/50">
						<div className="flex items-center gap-1">
							<span className="size-1.5 rounded-full bg-slate-400 animate-pulse" />
							<span className="size-1.5 rounded-full bg-slate-400 animate-pulse [animation-delay:200ms]" />
							<span className="size-1.5 rounded-full bg-slate-400 animate-pulse [animation-delay:400ms]" />
						</div>
					</div>
				)}

				{!isStreaming && (
					<div className="mt-2 flex items-center gap-3">
						{message.metadata && (
							<div className="flex items-center gap-1.5 text-xs text-slate-400">
								<span>{message.metadata.model}</span>
								<span>·</span>
								<span>{message.metadata.tokens} tokens</span>
								<span>·</span>
								<span>{message.metadata.latency}ms</span>
							</div>
						)}
						<div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
							<CopyButton text={content} />
							<Button
								variant="ghost"
								size="icon-xs"
								className="text-slate-400 hover:text-slate-600"
								onClick={() => resendMessage(message.id)}
							>
								<RefreshCw className="size-3.5" />
							</Button>
						</div>
					</div>
				)}
			</div>
		</div>
	);
}
