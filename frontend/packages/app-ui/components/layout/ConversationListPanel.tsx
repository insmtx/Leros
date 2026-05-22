"use client";

import type { Conversation } from "@leros/store";
import { useChatStore, useLayoutStore } from "@leros/store";
import { Button } from "@leros/ui/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@leros/ui/components/ui/dialog";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@leros/ui/components/ui/dropdown-menu";
import { ScrollArea } from "@leros/ui/components/ui/scroll-area";
import { cn } from "@leros/ui/lib/utils";
import { MoreHorizontal, Pencil, Plus, Search, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";

function formatConversationDate(timestamp: number) {
	const date = new Date(timestamp);
	const now = new Date();
	const isToday = date.toDateString() === now.toDateString();

	if (isToday) {
		return date.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
	}

	return date.toLocaleDateString("zh-CN", { month: "2-digit", day: "2-digit" });
}

function getConversationPreview(conv: Conversation) {
	if (conv.status === "active") return "正在进行的会话";
	if (conv.status === "archived") return "已归档";
	return conv.type === "assistant_instance" ? "AI 助手对话" : "暂无消息预览";
}

export function ConversationListPanel() {
	const {
		conversations,
		activeConversationId,
		conversationSearchQuery,
		conversationListOpen,
		switchConversation,
		createConversation,
		deleteConversation,
		updateConversationTitle,
		setConversationSearchQuery,
		fetchConversations,
	} = useLayoutStore((s) => s);

	const { setActiveSession, loadConversationMessages } = useChatStore((s) => s);

	// Rename dialog state
	const [renameDialogOpen, setRenameDialogOpen] = useState(false);
	const [renameTargetId, setRenameTargetId] = useState<string | null>(null);
	const [renameValue, setRenameValue] = useState("");

	useEffect(() => {
		fetchConversations();
	}, [fetchConversations]);

	const filteredConversations = conversationSearchQuery
		? conversations.filter((c) =>
				c.title.toLowerCase().includes(conversationSearchQuery.toLowerCase()),
			)
		: conversations;

	const handleConversationClick = (id: string) => {
		switchConversation(id);
		setActiveSession(id);
		loadConversationMessages(id);
	};

	const handleCreateConversation = async () => {
		const conv = await createConversation("新会话");
		if (conv) {
			switchConversation(conv.id);
			setActiveSession(conv.id);
			loadConversationMessages(conv.id);
		}
	};

	const handleDeleteConversation = async (id: string) => {
		await deleteConversation(id);
	};

	const handleOpenRename = (conv: { id: string; title: string }) => {
		setRenameTargetId(conv.id);
		setRenameValue(conv.title);
		setRenameDialogOpen(true);
	};

	const handleConfirmRename = async () => {
		if (renameTargetId && renameValue.trim()) {
			await updateConversationTitle(renameTargetId, renameValue.trim());
			setRenameDialogOpen(false);
			setRenameTargetId(null);
			setRenameValue("");
		}
	};

	if (!conversationListOpen) return null;

	return (
		<>
			<div
				data-slot="conversation-list-panel"
				className="flex h-full w-[288px] flex-col border-r border-slate-200/50 bg-white/95 transition-all duration-300"
			>
				<div className="flex h-14 items-center gap-2 border-b border-slate-200/50 px-3.5">
					<div className="relative min-w-0 flex-1">
						<Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-slate-400" />
						<input
							type="text"
							value={conversationSearchQuery}
							onChange={(e) => setConversationSearchQuery(e.target.value)}
							placeholder="搜索会话"
							className="w-full rounded-xl border-0 bg-slate-100/70 py-2 pl-7 pr-2 text-sm text-slate-600 placeholder:text-slate-400 ring-1 ring-transparent transition-colors focus:bg-white focus:outline-none focus:ring-blue-200"
						/>
					</div>
					<Button
						variant="ghost"
						size="icon-sm"
						className="shrink-0 rounded-xl text-slate-500 hover:bg-slate-50 hover:text-slate-700"
						onClick={handleCreateConversation}
					>
						<Plus className="size-4" />
					</Button>
				</div>

				<ScrollArea className="min-h-0 flex-1 overflow-hidden">
					<div className="px-2.5 py-3 pb-6">
						<div className="mb-2 flex items-center justify-between px-2">
							<span className="text-[11px] font-medium uppercase tracking-wide text-slate-400">
								最近会话
							</span>
							<span className="text-[11px] text-slate-300">{filteredConversations.length}</span>
						</div>
						{filteredConversations.map((conv) => (
							// biome-ignore lint/a11y/useSemanticElements: The row contains a nested menu button, so the row itself cannot be a button.
							<div
								key={conv.id}
								role="button"
								tabIndex={0}
								className={cn(
									"group relative mb-1 flex w-full cursor-pointer rounded-2xl px-4 py-3.5 text-left transition-all",
									activeConversationId === conv.id
										? "bg-slate-100 text-slate-900"
										: "text-slate-600 hover:bg-slate-50",
								)}
								onClick={() => handleConversationClick(conv.id)}
								onKeyDown={(e) => {
									if (e.key === "Enter" || e.key === " ") {
										e.preventDefault();
										handleConversationClick(conv.id);
									}
								}}
							>
								<div className="min-w-0 flex-1 pr-9">
									<span className="block truncate text-sm font-semibold">{conv.title}</span>
									<div className="mt-0.5 truncate text-xs text-slate-400">
										{getConversationPreview(conv)}
									</div>
								</div>
								<span className="absolute right-4 top-3.5 shrink-0 text-xs text-slate-400">
									{formatConversationDate(conv.updatedAt)}
								</span>
								<DropdownMenu>
									<DropdownMenuTrigger
										render={
											<Button
												variant="ghost"
												size="icon-xs"
												className="absolute bottom-2.5 right-3 shrink-0 text-slate-400 opacity-0 transition-opacity hover:text-slate-600 group-hover:opacity-100"
												onClick={(e: React.MouseEvent) => e.stopPropagation()}
											>
												<MoreHorizontal className="size-3.5" />
											</Button>
										}
									/>
									<DropdownMenuContent align="end" sideOffset={4}>
										<DropdownMenuItem onClick={() => handleOpenRename(conv)}>
											<Pencil className="size-3.5 mr-2" />
											<span>重命名</span>
										</DropdownMenuItem>
										<DropdownMenuItem
											variant="destructive"
											onClick={() => handleDeleteConversation(conv.id)}
										>
											<Trash2 className="size-3.5 mr-2" />
											<span>删除</span>
										</DropdownMenuItem>
									</DropdownMenuContent>
								</DropdownMenu>
							</div>
						))}
					</div>
				</ScrollArea>
			</div>

			{/* Rename Dialog */}
			<Dialog open={renameDialogOpen} onOpenChange={setRenameDialogOpen}>
				<DialogContent className="sm:max-w-md" showCloseButton={false}>
					<DialogHeader>
						<DialogTitle>重命名会话</DialogTitle>
						<DialogDescription>请输入新的会话名称</DialogDescription>
					</DialogHeader>
					<div className="mt-4">
						<input
							type="text"
							value={renameValue}
							onChange={(e) => setRenameValue(e.target.value)}
							onKeyDown={(e) => {
								if (e.key === "Enter") {
									handleConfirmRename();
								}
							}}
							placeholder="会话名称"
							autoFocus
							className="w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-slate-800 placeholder:text-slate-400 focus:border-blue-300 focus:outline-none transition-colors"
						/>
					</div>
					<DialogFooter className="mt-4">
						<Button variant="outline" onClick={() => setRenameDialogOpen(false)}>
							取消
						</Button>
						<button
							type="button"
							onClick={handleConfirmRename}
							disabled={!renameValue.trim()}
							className="inline-flex items-center justify-center rounded-lg bg-primary text-primary-foreground h-8 px-2.5 text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 hover:bg-primary/80"
						>
							确认
						</button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
}
