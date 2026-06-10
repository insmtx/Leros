"use client";

import type { RuntimeTodoItem, TodoStatus } from "@leros/store/types/chat";
import { cn } from "@leros/ui/lib/utils";
import { CheckCircle2, Circle, Loader2, XCircle } from "lucide-react";

const STATUS_LABEL: Record<TodoStatus, string> = {
	pending: "待开始",
	in_progress: "进行中",
	completed: "已完成",
	cancelled: "已取消",
};

export function TaskTodoProgressPanel({ todos }: { todos: RuntimeTodoItem[] }) {
	const total = todos.length;
	const completedCount = todos.filter((todo) => todo.status === "completed").length;
	const activeCount = todos.filter((todo) => todo.status === "in_progress").length;
	const progressPercent = total > 0 ? Math.round((completedCount / total) * 100) : 0;

	return (
		<div
			data-slot="task-todo-progress-panel"
			className="rounded-lg border border-[var(--leros-control-border)] bg-[var(--leros-surface)] shadow-sm"
		>
			<div className="border-b border-[var(--leros-control-border)] px-3.5 py-3">
				<div className="mb-2 flex items-center justify-between gap-2 text-xs">
					<span className="font-semibold text-[var(--leros-text-strong)]">
						{completedCount}/{total} 完成
					</span>
					{activeCount > 0 && (
						<span className="text-[var(--leros-primary)]">{activeCount} 进行中</span>
					)}
				</div>
				<div className="h-1.5 overflow-hidden rounded-full bg-[var(--leros-chat-control-bg)]">
					<div
						className="h-full rounded-full bg-[var(--leros-primary)] transition-[width] duration-300 ease-out"
						style={{ width: `${progressPercent}%` }}
					/>
				</div>
			</div>

			<ol className="px-3.5 py-3">
				{todos.map((todo, index) => (
					<TaskTodoProgressStep
						key={todo.id || `${todo.title}-${index}`}
						todo={todo}
						isLast={index === todos.length - 1}
					/>
				))}
			</ol>
		</div>
	);
}

function TaskTodoProgressStep({
	todo,
	isLast,
}: {
	todo: RuntimeTodoItem;
	isLast: boolean;
}) {
	const completed = todo.status === "completed";
	const cancelled = todo.status === "cancelled";

	return (
		<li className="relative flex gap-3 pb-4 last:pb-0">
			{!isLast && (
				<span
					aria-hidden
					className={cn(
						"absolute left-[11px] top-6 bottom-0 w-px",
						completed ? "bg-[var(--leros-primary-soft)]" : "bg-[var(--leros-control-border)]",
					)}
				/>
			)}
			<div className="relative z-[1] flex size-6 shrink-0 items-center justify-center rounded-full border border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)]">
				<TodoStatusIcon status={todo.status} />
			</div>
			<div className="min-w-0 flex-1 pt-0.5">
				<p
					className={cn(
						"text-sm font-medium leading-5 text-[var(--leros-text-strong)]",
						completed && "text-[var(--leros-text-muted)] line-through",
						cancelled && "text-[var(--leros-text-subtle)] line-through",
					)}
				>
					{todo.title}
				</p>
				<p className={cn("mt-0.5 text-xs", statusClassName(todo.status))}>
					{STATUS_LABEL[todo.status]}
				</p>
			</div>
		</li>
	);
}

function TodoStatusIcon({ status }: { status: TodoStatus }) {
	switch (status) {
		case "in_progress":
			return <Loader2 className="size-3.5 animate-spin text-[var(--leros-primary)]" />;
		case "completed":
			return <CheckCircle2 className="size-3.5 text-[var(--leros-primary)]" />;
		case "cancelled":
			return <XCircle className="size-3.5 text-[var(--leros-text-subtle)]" />;
		default:
			return <Circle className="size-3.5 text-[var(--leros-text-subtle)]" />;
	}
}

function statusClassName(status: TodoStatus) {
	switch (status) {
		case "in_progress":
		case "completed":
			return "text-[var(--leros-primary)]";
		case "cancelled":
			return "text-[var(--leros-text-subtle)]";
		default:
			return "text-[var(--leros-text-muted)]";
	}
}
