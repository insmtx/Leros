"use client";

import { useLayoutStore } from "@leros/store";
import { AssistantListView } from "../digitalAssistant/AssistantListView";
import { CenterCanvas } from "./CenterCanvas";
import { LeftRail } from "./LeftRail";
import { WorkbenchPanel } from "./WorkbenchPanel";

export function Shell({ logoSrc }: { logoSrc?: string }) {
	const currentView = useLayoutStore((s) => s.currentView);

	return (
		<div className="leros-app-shell">
			<LeftRail logoSrc={logoSrc} />
			{currentView === "chat" && <CenterCanvas />}
			{currentView === "workbench" && <WorkbenchPanel />}
			{currentView === "digitalAssistant" && <AssistantListView />}
		</div>
	);
}
