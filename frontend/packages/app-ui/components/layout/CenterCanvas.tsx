"use client";

import { useChatStore, useLayoutStore } from "@leros/store";
import { useLayoutEffect } from "react";
import { ChatHeader } from "../chat/ChatHeader";
import { MessageTimeline } from "../chat/MessageTimeline";
import { ChatInput } from "../input/ChatInput";

export function CenterCanvas() {
	const { resetLocalMessages } = useChatStore((s) => s);
	const { clearTaskDetailRoute } = useLayoutStore((s) => s);

	useLayoutEffect(() => {
		clearTaskDetailRoute();
		resetLocalMessages();
	}, [clearTaskDetailRoute, resetLocalMessages]);

	return (
		<div data-slot="center-canvas" className="flex h-full flex-1 flex-col bg-slate-50/80">
			<ChatHeader />
			<div className="flex min-h-0 flex-1 flex-col bg-[linear-gradient(180deg,#f8fafc_0%,#f6f7fb_100%)]">
				<MessageTimeline />
				<ChatInput />
			</div>
		</div>
	);
}
