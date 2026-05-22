"use client";

import type { RuntimeTodoItem, TodoStatus } from "@leros/store/types/chat";
import { cn } from "@leros/ui/lib/utils";
import { CheckCircle2, Circle, ListChecks, Loader2, XCircle } from "lucide-react";

const statusText: Record<TodoStatus, string> = {
	pending: "待开始",
	in_progress: "进行中",
	completed: "已完成",
	cancelled: "已取消",
};

export function TodoListBlock({ todos }: { todos: RuntimeTodoItem[] }) {
	const completedCount = todos.filter((todo) => todo.status === "completed").length;
	const activeCount = todos.filter((todo) => todo.status === "in_progress").length;

	return (
		<div
			data-slot="todo-list-block"
			className="max-w-[min(780px,92%)] rounded-lg border border-slate-200 bg-white/90 shadow-sm"
		>
			<div className="flex items-center justify-between gap-3 border-b border-slate-200 px-3 py-2">
				<div className="flex min-w-0 items-center gap-2">
					<ListChecks className="size-4 shrink-0 text-blue-500" />
					<span className="truncate text-sm font-medium text-slate-700">任务进度</span>
				</div>
				<div className="flex shrink-0 items-center gap-2 text-xs text-slate-500">
					{activeCount > 0 && <span className="text-blue-600">{activeCount} 进行中</span>}
					<span>
						{completedCount}/{todos.length} 完成
					</span>
				</div>
			</div>

			<div className="space-y-1 px-3 py-2">
				{todos.map((todo, index) => (
					<TodoListItem key={todo.id || index} todo={todo} />
				))}
			</div>
		</div>
	);
}

function TodoListItem({ todo }: { todo: RuntimeTodoItem }) {
	const completed = todo.status === "completed";
	const cancelled = todo.status === "cancelled";

	return (
		<div className="flex items-center gap-2 rounded-md px-1.5 py-1.5 text-sm">
			<TodoStatusIcon status={todo.status} />
			<div className="min-w-0 flex-1">
				<div
					className={cn(
						"truncate text-slate-700",
						completed && "text-slate-500 line-through",
						cancelled && "text-slate-400 line-through",
					)}
				>
					{todo.title}
				</div>
			</div>
			{todo.priority && (
				<span className="shrink-0 rounded border border-slate-200 px-1.5 py-0.5 text-[11px] leading-none text-slate-500">
					{todo.priority}
				</span>
			)}
			<span className={cn("shrink-0 text-xs", statusClassName(todo.status))}>
				{statusText[todo.status]}
			</span>
		</div>
	);
}

function TodoStatusIcon({ status }: { status: TodoStatus }) {
	switch (status) {
		case "in_progress":
			return <Loader2 className="size-3.5 shrink-0 animate-spin text-blue-500" />;
		case "completed":
			return <CheckCircle2 className="size-3.5 shrink-0 text-green-500" />;
		case "cancelled":
			return <XCircle className="size-3.5 shrink-0 text-slate-400" />;
		default:
			return <Circle className="size-3.5 shrink-0 text-slate-300" />;
	}
}

function statusClassName(status: TodoStatus) {
	switch (status) {
		case "in_progress":
			return "text-blue-600";
		case "completed":
			return "text-green-600";
		case "cancelled":
			return "text-slate-400";
		default:
			return "text-slate-500";
	}
}
