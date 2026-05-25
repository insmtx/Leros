"use client";

import type { Project, ProjectArtifact, ProjectTask } from "@leros/store";
import { useLayoutStore } from "@leros/store";
import { cn } from "@leros/ui/lib/utils";
import {
	Bot,
	CheckCircle2,
	FileImage,
	FileText,
	LayoutPanelLeft,
	Search,
	Settings,
	Table2,
} from "lucide-react";
import { MessageTimeline } from "../chat/MessageTimeline";
import { ChatInput } from "../input/ChatInput";

const projectTabs = [
	{ id: "chat" as const, label: "会话" },
	{ id: "tasks" as const, label: "任务" },
	{ id: "files" as const, label: "文件" },
	{ id: "memory" as const, label: "记忆" },
];

export function ProjectPage() {
	const { projects, activeProjectId, activeProjectTab, switchView, setActiveProjectTab } =
		useLayoutStore((s) => s);

	const project = projects.find((item) => item.id === activeProjectId) ?? projects[0];

	if (!project) {
		return (
			<div className="flex h-full flex-1 items-center justify-center bg-[var(--leros-app-bg)] text-[var(--leros-text-muted)]">
				暂无项目
			</div>
		);
	}

	return (
		<div data-slot="project-page" className="flex h-full flex-1 flex-col bg-[var(--leros-surface)]">
			<header className="flex h-16 shrink-0 items-center justify-between border-b border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] px-10">
				<div className="flex items-center gap-3 text-[var(--leros-text-muted)]">
					<button
						type="button"
						onClick={() => switchView("workbench")}
						className="text-xs font-semibold uppercase tracking-widest hover:text-[var(--leros-text-strong)]"
					>
						Projects
					</button>
					<span className="text-[var(--leros-text-subtle)]">/</span>
					<h1 className="text-base font-bold text-[var(--leros-text-strong)]">{project.name}</h1>
				</div>
				<div className="flex items-center gap-6 text-[var(--leros-text)]">
					<button
						type="button"
						className="rounded-full p-1.5 transition-colors hover:bg-[var(--leros-primary-softer)]"
					>
						<Search className="size-5" />
					</button>
					<button
						type="button"
						className="rounded-full p-1.5 transition-colors hover:bg-[var(--leros-primary-softer)]"
					>
						<LayoutPanelLeft className="size-5" />
					</button>
					<button
						type="button"
						className="rounded-full p-1.5 transition-colors hover:bg-[var(--leros-primary-softer)]"
					>
						<Settings className="size-5" />
					</button>
				</div>
			</header>

			<nav className="flex h-[48px] shrink-0 items-end gap-8 border-b border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] px-10">
				{projectTabs.map((tab) => (
					<button
						key={tab.id}
						type="button"
						onClick={() => setActiveProjectTab(tab.id)}
						className={cn(
							"relative h-full px-1 pb-2 text-sm font-semibold transition-colors",
							activeProjectTab === tab.id
								? "text-[var(--leros-primary)]"
								: "text-[var(--leros-text-muted)] hover:text-[var(--leros-text-strong)]",
						)}
					>
						{tab.label}
						{activeProjectTab === tab.id && (
							<span className="absolute bottom-0 left-0 h-0.5 w-full rounded-full bg-[var(--leros-primary)]" />
						)}
					</button>
				))}
			</nav>

			<div className="min-h-0 flex flex-1">
				<main
					className={cn(
						"min-w-0 flex-1",
						activeProjectTab === "chat"
							? "flex min-h-0 flex-col bg-[var(--leros-surface)]"
							: "overflow-y-auto px-10 py-8",
					)}
				>
					{activeProjectTab === "chat" && <ProjectChat />}
					{activeProjectTab === "tasks" && <ProjectTasks tasks={project.tasks} />}
					{activeProjectTab === "files" && <ProjectFiles files={project.files} />}
					{activeProjectTab === "memory" && <ProjectMemories project={project} />}
				</main>

				<aside className="flex w-[300px] shrink-0 flex-col border-l border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] px-5 py-6">
					<div className="min-h-0 flex-1 space-y-8 overflow-y-auto pr-1">
						<section>
							<div className="mx-auto mb-4 flex w-full max-w-[250px] items-center justify-between">
								<h2 className="text-xs font-semibold text-[var(--leros-text-muted)]">任务</h2>
								<span className="rounded-md bg-[var(--leros-primary-soft)] px-2 py-0.5 text-xs font-semibold text-[var(--leros-primary)]">
									{project.tasks.length} 项
								</span>
							</div>
							<ProjectTaskList tasks={project.tasks} compact />
						</section>

						<section>
							<div className="mx-auto mb-4 flex w-full max-w-[250px] items-center justify-between">
								<h2 className="text-xs font-semibold text-[var(--leros-text-muted)]">产物</h2>
								<span className="rounded-md bg-[var(--leros-chat-control-bg)] px-2 py-0.5 text-xs font-semibold text-[var(--leros-text)]">
									{project.artifacts.length} 个
								</span>
							</div>
							<ProjectArtifactList artifacts={project.artifacts} compact />
						</section>
					</div>
				</aside>
			</div>
		</div>
	);
}

function ProjectChat() {
	return (
		<div className="flex min-h-0 flex-1 flex-col">
			<MessageTimeline
				emptyState={<ProjectEmptyState />}
				contentClassName="max-w-[780px] px-8 py-8 sm:px-8 lg:px-8"
			/>
			<ChatInput variant="project" />
		</div>
	);
}

function ProjectEmptyState() {
	return (
		<div className="flex h-full items-center justify-center px-8">
			<div className="flex max-w-[320px] flex-col items-center text-center">
				<div className="flex size-12 items-center justify-center rounded-full bg-[var(--leros-primary-softer)] text-[var(--leros-primary)]">
					<Bot className="size-6" />
				</div>
				<h2 className="mt-5 text-lg font-semibold text-[var(--leros-text-strong)]">开始项目会话</h2>
				<p className="mt-2 text-sm leading-6 text-[var(--leros-text-muted)]">
					把需求、问题或上下文发给 AI，后续讨论会沉淀在当前项目中。
				</p>
			</div>
		</div>
	);
}

function ProjectTasks({ tasks }: { tasks: ProjectTask[] }) {
	return (
		<div className="mx-auto w-full max-w-[720px]">
			<h2 className="text-lg font-semibold text-[var(--leros-text-strong)]">任务</h2>
			<div className="mt-4">
				<ProjectTaskList tasks={tasks} />
			</div>
		</div>
	);
}

function ProjectTaskList({ tasks, compact = false }: { tasks: ProjectTask[]; compact?: boolean }) {
	return (
		<div className={cn("w-full", compact ? "mx-auto max-w-[250px] space-y-3" : "space-y-3")}>
			{tasks.map((task) => (
				<div
					key={task.id}
					className={cn(
						"flex items-start border border-[var(--leros-control-border)] bg-[var(--leros-surface)] shadow-sm",
						compact ? "gap-3 rounded-lg px-3.5 py-3" : "gap-3.5 rounded-lg px-4 py-3.5",
					)}
				>
					<CheckCircle2
						className={cn(
							"mt-0.5 size-4 shrink-0",
							task.status === "done"
								? "text-[var(--leros-primary)]"
								: "text-[var(--leros-text-muted)]",
						)}
					/>
					<div className="min-w-0">
						<div className="truncate text-sm font-semibold leading-5 text-[var(--leros-text-strong)]">
							{task.title}
						</div>
						<div className="mt-1 truncate text-xs leading-4 text-[var(--leros-text-muted)]">
							{task.meta}
						</div>
					</div>
				</div>
			))}
		</div>
	);
}

function ProjectFiles({ files }: { files: ProjectArtifact[] }) {
	return (
		<div className="mx-auto w-full max-w-[720px]">
			<h2 className="text-lg font-semibold text-[var(--leros-text-strong)]">文件</h2>
			<div className="mt-4">
				<ProjectArtifactList artifacts={files} emptyText="暂无文件" />
			</div>
		</div>
	);
}

function ProjectMemories({ project }: { project: Project }) {
	return (
		<div className="mx-auto w-full max-w-[720px]">
			<h2 className="text-lg font-semibold text-[var(--leros-text-strong)]">记忆</h2>
			<div className="mt-4 space-y-3">
				{project.memories.map((memory) => (
					<div
						key={memory.id}
						className="rounded-lg border border-[var(--leros-control-border)] bg-[var(--leros-surface)] px-4 py-3.5 shadow-sm"
					>
						<div className="text-sm font-semibold leading-5 text-[var(--leros-text-strong)]">
							{memory.title}
						</div>
						<p className="mt-1.5 text-xs leading-5 text-[var(--leros-text-muted)]">
							{memory.content}
						</p>
					</div>
				))}
				{project.memories.length === 0 && (
					<div className="rounded-lg border border-dashed border-[var(--leros-control-border)] px-4 py-8 text-center text-xs text-[var(--leros-text-muted)]">
						暂无记忆
					</div>
				)}
			</div>
		</div>
	);
}

function ProjectArtifactList({
	artifacts,
	emptyText = "暂无产物",
	compact = false,
}: {
	artifacts: ProjectArtifact[];
	emptyText?: string;
	compact?: boolean;
}) {
	if (artifacts.length === 0) {
		return (
			<div className="rounded-lg border border-dashed border-[var(--leros-control-border)] px-4 py-8 text-center text-xs text-[var(--leros-text-muted)]">
				{emptyText}
			</div>
		);
	}

	return (
		<div className={cn("w-full", compact ? "mx-auto max-w-[250px] space-y-3" : "space-y-3")}>
			{artifacts.map((artifact) => (
				<div
					key={artifact.id}
					className={cn(
						"flex items-center border border-[var(--leros-control-border)] bg-[var(--leros-surface)] shadow-sm",
						compact ? "gap-3 rounded-lg px-3.5 py-3" : "gap-3.5 rounded-lg px-4 py-3.5",
					)}
				>
					<div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-[var(--leros-primary-softer)] text-[var(--leros-text)]">
						<ArtifactIcon type={artifact.type} />
					</div>
					<div className="min-w-0">
						<div className="truncate text-sm font-semibold leading-5 text-[var(--leros-text-strong)]">
							{artifact.name}
						</div>
						<div className="mt-1 truncate text-xs leading-4 text-[var(--leros-text-muted)]">
							{artifact.size} · {artifact.updatedAt}
						</div>
					</div>
				</div>
			))}
		</div>
	);
}

function ArtifactIcon({ type }: { type: ProjectArtifact["type"] }) {
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
