"use client";

import type { NavItem } from "@singeros/store";
import { useLayoutStore } from "@singeros/store";
import { Button } from "@singeros/ui/components/ui/button";
import { ScrollArea } from "@singeros/ui/components/ui/scroll-area";
import { cn } from "@singeros/ui/lib/utils";
import {
	BookOpen,
	Bot,
	Calendar,
	ChevronDown,
	ChevronLeft,
	ChevronRight,
	Code2,
	GitBranch,
	Hammer,
	MessageSquare,
	Network,
	Paintbrush,
	Plus,
	Settings,
	Star,
	Terminal,
	Users,
} from "lucide-react";

const iconMap: Record<string, React.ReactNode> = {
	IconRobot: <Bot className="size-4" />,
	IconCommand: <Terminal className="size-4" />,
	IconUsers: <Users className="size-4" />,
	IconBook: <BookOpen className="size-4" />,
	IconStar: <Star className="size-4" />,
	IconGitBranch: <GitBranch className="size-4" />,
	IconCode: <Code2 className="size-4" />,
	IconHammer: <Hammer className="size-4" />,
	IconPaint: <Paintbrush className="size-4" />,
	IconNetwork: <Network className="size-4" />,
	IconReport: <Calendar className="size-4" />,
	IconCalendar: <Calendar className="size-4" />,
	IconSettings2: <Settings className="size-4" />,
	IconMessage: <MessageSquare className="size-4" />,
};

export function LeftRail() {
	const {
		leftRailCollapsed,
		navGroups,
		collapsedNavGroups,
		conversationListOpen,
		toggleLeftRail,
		toggleNavGroup,
		toggleConversationList,
	} = useLayoutStore((s) => s);

	return (
		<div
			className={cn(
				"flex h-full flex-col border-r border-slate-200 bg-white transition-all duration-300",
				leftRailCollapsed ? "w-[52px]" : "w-[260px]",
			)}
		>
			<div className="flex h-12 items-center justify-between border-b border-slate-200 px-4">
				{!leftRailCollapsed && (
					<h2 className="text-sm font-medium tracking-wide uppercase text-slate-600">导航</h2>
				)}
				<button
					type="button"
					onClick={toggleLeftRail}
					className={cn(
						"flex items-center justify-center rounded-md p-1 text-slate-400 hover:text-slate-600 hover:bg-slate-50 transition-colors",
						leftRailCollapsed ? "mx-auto" : "ml-auto",
					)}
				>
					{leftRailCollapsed ? (
						<ChevronRight className="size-4" />
					) : (
						<ChevronLeft className="size-4" />
					)}
				</button>
			</div>

			<ScrollArea className="flex-1">
				<div className="p-1.5">
					{navGroups.map((group) => {
						const isCollapsed = collapsedNavGroups.has(group.id);

						if (leftRailCollapsed) {
							return (
								<div key={group.id} className="mb-1">
									{group.items.map((item: NavItem) => (
										<CollapsedNavItemButton key={item.id} item={item} />
									))}
								</div>
							);
						}

						return (
							<div key={group.id} className="mb-0.5">
								{group.label && (
									<button
										type="button"
										onClick={() => toggleNavGroup(group.id)}
										className="flex w-full items-center gap-1 rounded-md px-2 py-1.5 text-xs font-medium text-slate-500 hover:bg-slate-50 transition-colors"
									>
										{isCollapsed ? (
											<ChevronRight className="size-3.5" />
										) : (
											<ChevronDown className="size-3.5" />
										)}
										<span className="tracking-wide uppercase">{group.label}</span>
									</button>
								)}

								{!isCollapsed && (
									<div className={cn("mt-0.5", group.label && "ml-2")}>
										{group.items.map((item: NavItem) =>
											item.id === "ai-assistant" ? (
												<button
													key={item.id}
													type="button"
													onClick={() => toggleConversationList()}
													className={cn(
														"group flex items-center gap-2.5 rounded-md px-2.5 py-2 text-sm cursor-pointer transition-colors w-full text-left",
														conversationListOpen
															? "bg-blue-50 text-blue-700"
															: "text-slate-600 hover:bg-slate-50 hover:text-slate-800",
													)}
												>
													{iconMap[item.icon]}
													<span className="truncate">{item.label}</span>
												</button>
											) : (
												<NavItemButton key={item.id} item={item} active={false} />
											),
										)}
									</div>
								)}
							</div>
						);
					})}
				</div>
			</ScrollArea>

			{!leftRailCollapsed && (
				<div className="border-t border-slate-200 p-2">
					<Button variant="ghost" size="sm" className="w-full justify-start text-slate-500">
						<Plus className="size-4 mr-1.5" />
						新建会话
					</Button>
				</div>
			)}
		</div>
	);
}

function NavItemButton({ item, active }: { item: NavItem; active: boolean }) {
	const icon = iconMap[item.icon];
	return (
		<button
			type="button"
			className={cn(
				"group flex items-center gap-2.5 rounded-md px-2.5 py-2 text-sm cursor-pointer transition-colors w-full text-left",
				active
					? "bg-blue-50 text-blue-700"
					: "text-slate-600 hover:bg-slate-50 hover:text-slate-800",
			)}
		>
			{icon}
			<span className="truncate">{item.label}</span>
			{item.badge && (
				<span className="ml-auto rounded-full bg-red-100 text-red-600 px-1.5 py-0.5 text-xs">
					{item.badge}
				</span>
			)}
		</button>
	);
}

function CollapsedNavItemButton({ item }: { item: NavItem }) {
	const icon = iconMap[item.icon];
	return (
		<button
			type="button"
			className="flex items-center justify-center rounded-md p-2 text-slate-500 hover:bg-slate-50 hover:text-slate-700 transition-colors w-full cursor-pointer"
			title={item.label}
		>
			{icon}
		</button>
	);
}
