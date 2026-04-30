"use client";

import { ChatHeader } from "../chat/ChatHeader";
import { MessageTimeline } from "../chat/MessageTimeline";
import { ChatInput } from "../input/ChatInput";

export function CenterCanvas() {
	return (
		<div data-slot="center-canvas" className="flex h-full flex-1 flex-col bg-slate-50">
			<ChatHeader />
			<MessageTimeline />
			<ChatInput />
		</div>
	);
}
