"use client";

import type { ProjectArtifact, ProjectTask } from "@leros/store";
import { mapBackendArtifactToProjectArtifact, useChatStore, useLayoutStore } from "@leros/store";
import { artifactApi } from "@leros/store/api/artifactApi";
import { taskApi } from "@leros/store/api/taskApi";
import { cn } from "@leros/ui/lib/utils";
import {
	ArrowLeft,
	Bot,
	Calendar,
	CheckCircle2,
	Circle,
	FileImage,
	FileText,
	LoaderCircle,
	Table2,
	Tag,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { MessageTimeline } from "../chat/MessageTimeline";
import { ChatInput } from "../input/ChatInput";
import { ArtifactPreviewDialog } from "./ArtifactPreviewDialog";
import type { AppNavigation } from "./LeftRail";

const STATUS_LABEL: Record<string, string> = {
	todo: "待办",
	in_progress: "进行中",
	done: "已完成",
};

export function TaskDetailPage({
	projectId,
	taskId,
	sessionId,
	navigation,
}: {
	projectId?: string;
	taskId?: string;
	sessionId?: string | null;
	navigation?: AppNavigation;
}) {
	const {
		activeTaskDetailProjectId,
		activeTaskDetailTaskId,
		activeTaskDetailSessionId,
		projects,
		fetchProjects,
		setTaskDetailRoute,
		switchView,
		switchProject,
	} = useLayoutStore((s) => s);

	const {
		activeSessionId,
		isGenerating,
		setActiveSession,
		loadConversationMessages,
	} = useChatStore((s) => s);

	const [task, setTask] = useState<ProjectTask | null>(null);
	const [artifacts, setArtifacts] = useState<ProjectArtifact[]>([]);
	const [previewArtifact, setPreviewArtifact] = useState<ProjectArtifact | null>(null);

	const resolvedProjectId = projectId ?? activeTaskDetailProjectId;
	const resolvedTaskId = taskId ?? activeTaskDetailTaskId;
	const resolvedSessionId = sessionId ?? activeTaskDetailSessionId;
	const project = projects.find((p) => p.id === resolvedProjectId);

	const fetchArtifacts = useCallback(async (taskId: string) => {
		try {
			const res = await artifactApi.listTaskArtifacts(taskId);
			setArtifacts((res.data.data ?? []).map(mapBackendArtifactToProjectArtifact));
		} catch (err) {
			console.error("TaskDetailPage fetch artifacts error:", err);
			setArtifacts([]);
		}
	}, []);

	useEffect(() => {
		fetchProjects();
	}, [fetchProjects]);

	useEffect(() => {
		if (!projectId || !taskId) return;
		setTaskDetailRoute(projectId, taskId, sessionId ?? null);
	}, [projectId, taskId, sessionId, setTaskDetailRoute]);

	useEffect(() => {
		if (!resolvedSessionId) return;

		setActiveSession(resolvedSessionId);
		if (activeSessionId === resolvedSessionId && isGenerating) return;
		loadConversationMessages(resolvedSessionId);
	}, [
		resolvedSessionId,
		activeSessionId,
		isGenerating,
		setActiveSession,
		loadConversationMessages,
	]);

	useEffect(() => {
		if (!resolvedTaskId) return;

		taskApi
			.get({ public_id: resolvedTaskId })
			.then((res) => {
				const bt = res.data.data;
				if (bt) {
					setTask({
						id: bt.public_id,
						title: bt.title,
						meta: bt.description ?? bt.task_type ?? "",
						status: (bt.status as ProjectTask["status"]) ?? "todo",
						taskType: bt.task_type,
						deadline: bt.deadline,
						description: bt.description,
					});
				}
			})
			.catch((err) => {
				console.error("TaskDetailPage fetch task error:", err);
			});

		fetchArtifacts(resolvedTaskId);
	}, [resolvedTaskId, fetchArtifacts]);

	useEffect(() => {
		if (!resolvedTaskId || isGenerating) return;
		fetchArtifacts(resolvedTaskId);
	}, [resolvedTaskId, fetchArtifacts, isGenerating]);

	if (!resolvedProjectId || !resolvedTaskId) {
		return (
			<div className="flex h-full flex-1 items-center justify-center bg-[var(--leros-app-bg)] text-[var(--leros-text-muted)]">
				无任务详情
			</div>
		);
	}

	return (
		<div
			data-slot="task-detail-page"
			className="flex h-full flex-1 flex-col bg-[var(--leros-surface)]"
		>
			<header className="flex h-16 shrink-0 items-center justify-between border-b border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] px-10">
				<div className="flex items-center gap-3 text-[var(--leros-text-muted)]">
					{project && (
						<>
							<button
								type="button"
								onClick={() => {
									if (navigation && resolvedProjectId) {
										navigation.goToProject(resolvedProjectId);
										return;
									}
									resolvedProjectId && switchProject(resolvedProjectId);
								}}
								className="text-xs font-semibold uppercase tracking-widest hover:text-[var(--leros-text-strong)]"
							>
								{project.name}
							</button>
							<span className="text-[var(--leros-text-subtle)]">/</span>
						</>
					)}
					<h1 className="text-base font-bold text-[var(--leros-text-strong)]">
						{task?.title ?? "任务"}
					</h1>
				</div>
				<div className="flex items-center gap-3">
					<button
						type="button"
						onClick={() => {
							if (navigation) {
								navigation.goToRoute("workbench");
								return;
							}
							switchView("workbench");
						}}
						className="flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-semibold text-[var(--leros-text-muted)] transition-colors hover:bg-[var(--leros-primary-softer)] hover:text-[var(--leros-primary)]"
					>
						<ArrowLeft className="size-3.5" />
						返回工作台
					</button>
				</div>
			</header>

			{task && (
				<div className="shrink-0 border-b border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] px-10 py-4">
					<div className="flex flex-wrap items-center gap-4">
						<span
							className={cn(
								"inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold",
								task.status === "done"
									? "bg-[var(--leros-primary-soft)] text-[var(--leros-primary)]"
									: task.status === "in_progress"
										? "bg-[var(--leros-warning)]/10 text-[var(--leros-warning)]"
										: "bg-[var(--leros-chat-control-bg)] text-[var(--leros-text-muted)]",
							)}
						>
							{task.status === "done" ? (
								<CheckCircle2 className="size-3.5" />
							) : task.status === "in_progress" ? (
								<LoaderCircle className="size-3.5" />
							) : (
								<Circle className="size-3.5" />
							)}
							{STATUS_LABEL[task.status] ?? task.status}
						</span>
						{task.taskType && (
							<span className="inline-flex items-center gap-1 rounded-full bg-[var(--leros-primary-softer)] px-2.5 py-0.5 text-xs font-medium text-[var(--leros-primary)]">
								<Tag className="size-3" />
								{task.taskType}
							</span>
						)}
						{task.deadline && (
							<span className="inline-flex items-center gap-1 rounded-full bg-[var(--leros-chat-control-bg)] px-2.5 py-0.5 text-xs font-medium text-[var(--leros-text-muted)]">
								<Calendar className="size-3" />
								{task.deadline}
							</span>
						)}
					</div>
				</div>
			)}

			<div className="min-h-0 flex flex-1">
				<main className="flex min-h-0 flex-1 flex-col">
					<MessageTimeline
						emptyState={<TaskChatEmptyState />}
						contentClassName="max-w-[780px] px-8 py-8 sm:px-8 lg:px-8"
					/>
					<ChatInput variant="project" />
				</main>

				<aside className="flex w-[320px] shrink-0 flex-col border-l border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] px-5 py-6">
					<div className="min-h-0 flex-1 space-y-8 overflow-y-auto pr-1">
						{task?.description && (
							<section>
								<h3 className="mb-3 text-xs font-semibold text-[var(--leros-text-muted)]">
									任务描述
								</h3>
								<p className="text-sm leading-relaxed text-[var(--leros-text)]">
									{task.description}
								</p>
							</section>
						)}
						{project && (
							<section>
								<h3 className="mb-3 text-xs font-semibold text-[var(--leros-text-muted)]">
									所属项目
								</h3>
								<div className="rounded-lg border border-[var(--leros-control-border)] bg-[var(--leros-surface)] p-3.5">
									<p className="text-sm font-semibold text-[var(--leros-text-strong)]">
										{project.name}
									</p>
									{project.description && (
										<p className="mt-1 text-xs text-[var(--leros-text-muted)]">
											{project.description}
										</p>
									)}
								</div>
							</section>
						)}
						<section>
							<div className="mb-3 flex items-center justify-between">
								<h3 className="text-xs font-semibold text-[var(--leros-text-muted)]">任务产物</h3>
								<span className="rounded-md bg-[var(--leros-chat-control-bg)] px-2 py-0.5 text-xs font-semibold text-[var(--leros-text)]">
									{artifacts.length} 个
								</span>
							</div>
							<TaskArtifactList artifacts={artifacts} onPreview={setPreviewArtifact} />
						</section>
					</div>
				</aside>
			</div>
			<ArtifactPreviewDialog
				artifact={previewArtifact}
				open={previewArtifact !== null}
				onOpenChange={(open) => {
					if (!open) setPreviewArtifact(null);
				}}
			/>
		</div>
	);
}

function TaskChatEmptyState() {
	return (
		<div className="flex h-full items-center justify-center px-8">
			<div className="flex max-w-[320px] flex-col items-center text-center">
				<div className="flex size-12 items-center justify-center rounded-full bg-[var(--leros-primary-softer)] text-[var(--leros-primary)]">
					<Bot className="size-6" />
				</div>
				<h2 className="mt-5 text-lg font-semibold text-[var(--leros-text-strong)]">任务会话</h2>
				<p className="mt-2 text-sm leading-6 text-[var(--leros-text-muted)]">
					在此与 AI 协作完成任务讨论，发送消息即可开始对话。
				</p>
			</div>
		</div>
	);
}

function TaskArtifactList({
	artifacts,
	onPreview,
}: {
	artifacts: ProjectArtifact[];
	onPreview: (artifact: ProjectArtifact) => void;
}) {
	if (artifacts.length === 0) {
		return (
			<div className="rounded-lg border border-dashed border-[var(--leros-control-border)] px-4 py-8 text-center text-xs text-[var(--leros-text-muted)]">
				暂无产物
			</div>
		);
	}

	return (
		<div className="space-y-3">
			{artifacts.map((artifact) => (
				<button
					type="button"
					key={artifact.id}
					onClick={() => onPreview(artifact)}
					className="flex w-full items-center gap-3 rounded-lg border border-[var(--leros-control-border)] bg-[var(--leros-surface)] px-3.5 py-3 text-left shadow-sm transition-colors hover:border-[var(--leros-primary-soft)] hover:bg-[var(--leros-primary-softer)]/35"
					title="预览产物"
				>
					<div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-[var(--leros-primary-softer)] text-[var(--leros-text)]">
						<TaskArtifactIcon type={artifact.type} />
					</div>
					<div className="min-w-0">
						<div className="truncate text-sm font-semibold leading-5 text-[var(--leros-text-strong)]">
							{artifact.name}
						</div>
						<div className="mt-1 truncate text-xs leading-4 text-[var(--leros-text-muted)]">
							{artifact.size}
						</div>
					</div>
				</button>
			))}
		</div>
	);
}

function TaskArtifactIcon({ type }: { type: ProjectArtifact["type"] }) {
	const className = "size-4";

	switch (type) {
		case "spreadsheet":
			return <Table2 className={className} />;
		case "image":
			return <FileImage className={className} />;
		default:
			return <FileText className={className} />;
	}
}
