"use client";

import { formatTime } from "@leros/store";
import type { Message } from "@leros/store/types/chat";
import { Button } from "@leros/ui/components/ui/button";
import { Check, Copy } from "lucide-react";
import { useState } from "react";

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
			className={
				copied
					? "text-green-400"
					: "text-slate-300 opacity-0 group-hover:opacity-100 transition-opacity hover:text-slate-400"
			}
			onClick={handleCopy}
		>
			{copied ? <Check className="size-3.5" /> : <Copy className="size-3.5" />}
		</Button>
	);
}

export function UserMessageBubble({ message }: { message: Message }) {
	return (
		<div data-slot="user-message" className="flex justify-end group">
			<div className="flex max-w-[min(720px,78%)] flex-col items-end">
				<div className="mb-1.5 flex items-center gap-2">
					<CopyButton text={message.content} />
					<span className="text-xs text-slate-400">{formatTime(message.timestamp)}</span>
					<span className="text-xs font-medium text-slate-500">你</span>
				</div>
				<div className="w-fit rounded-2xl rounded-tr-md bg-blue-600 px-4 py-3 text-sm leading-7 text-white shadow-sm shadow-blue-600/10">
					{message.content}
				</div>
			</div>
		</div>
	);
}
