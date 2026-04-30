"use client";

import { Badge } from "@singeros/ui/components/ui/badge";
import { Button } from "@singeros/ui/components/ui/button";
import { Bell, ChevronDown, User } from "lucide-react";

export function TopBar() {
	return (
		<div
			data-slot="topbar"
			className="flex h-12 items-center justify-between border-b border-slate-200 bg-white px-4"
		>
			<div className="flex items-center gap-3">
				<span className="text-sm font-semibold text-slate-900 tracking-tight">SingerOS</span>
				<Badge variant="secondary" className="text-xs">
					v0.1
				</Badge>
			</div>

			<div className="flex items-center gap-3">
				<div className="flex items-center gap-1.5">
					<span className="relative flex size-2">
						<span className="absolute inline-flex size-full rounded-full bg-green-400 opacity-75 animate-ping" />
						<span className="relative inline-flex size-2 rounded-full bg-green-500" />
					</span>
					<span className="text-xs text-slate-500">AI 在线</span>
				</div>

				<Button variant="ghost" size="icon-sm" className="text-slate-500 hover:text-slate-700">
					<Bell className="size-4" />
				</Button>

				<button
					type="button"
					className="flex items-center gap-2 rounded-md hover:bg-slate-50 px-2 py-1 transition-colors"
				>
					<div className="size-7 rounded-full bg-blue-500 flex items-center justify-center text-white text-xs font-medium">
						<User className="size-4" />
					</div>
					<span className="text-sm text-slate-700">用户</span>
					<ChevronDown className="size-3 text-slate-400" />
				</button>
			</div>
		</div>
	);
}
