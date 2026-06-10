"use client";

import {
	Command,
	CommandEmpty,
	CommandGroup,
	CommandItem,
	CommandList,
} from "@leros/ui/components/ui/command";
import { cn } from "@leros/ui/lib/utils";
import { Bot } from "lucide-react";
import {
	forwardRef,
	useCallback,
	useEffect,
	useImperativeHandle,
	useMemo,
	useRef,
	useState,
} from "react";
import { type ChatCommand, mockAssistants, mockChatCommands } from "./mockDirectiveData";

type DirectiveKind = "assistant" | "command";

type DirectiveToken = {
	id: string;
	kind: DirectiveKind;
	start: number;
	end: number;
	label: string;
};

type ActiveTrigger = {
	kind: DirectiveKind;
	start: number;
	end: number;
	query: string;
};

type AssistantOption = {
	code: string;
	name: string;
	description: string;
};

export type StructuredComposerHandle = {
	openAssistantPicker: () => void;
};

type StructuredComposerProps = {
	value: string;
	onChange: (value: string) => void;
	onSubmit: () => void;
	onPaste: (event: React.ClipboardEvent<HTMLTextAreaElement>) => void;
	onFocus: () => void;
	onBlur: () => void;
	placeholder: string;
	isProjectVariant: boolean;
};

function findTrigger(value: string, cursor: number): ActiveTrigger | null {
	const prefix = value.slice(0, cursor);
	const assistantMatch = prefix.match(/(?:^|\s)@([^\s@/]*)$/);
	if (assistantMatch) {
		const query = assistantMatch[1] ?? "";
		return {
			kind: "assistant",
			start: cursor - query.length - 1,
			end: cursor,
			query,
		};
	}

	const commandMatch = prefix.match(/(?:^|\s)\/([^\s@/]*)$/);
	if (commandMatch) {
		const query = commandMatch[1] ?? "";
		return {
			kind: "command",
			start: cursor - query.length - 1,
			end: cursor,
			query,
		};
	}

	return null;
}

function updateTokensForTextChange(
	tokens: DirectiveToken[],
	previousValue: string,
	nextValue: string,
): DirectiveToken[] {
	let prefixLength = 0;
	while (
		prefixLength < previousValue.length &&
		prefixLength < nextValue.length &&
		previousValue[prefixLength] === nextValue[prefixLength]
	) {
		prefixLength += 1;
	}

	let suffixLength = 0;
	while (
		suffixLength < previousValue.length - prefixLength &&
		suffixLength < nextValue.length - prefixLength &&
		previousValue[previousValue.length - suffixLength - 1] ===
			nextValue[nextValue.length - suffixLength - 1]
	) {
		suffixLength += 1;
	}

	const previousEditEnd = previousValue.length - suffixLength;
	const delta = nextValue.length - previousValue.length;

	return tokens
		.flatMap((token) => {
			if (previousEditEnd <= token.start) {
				return [{ ...token, start: token.start + delta, end: token.end + delta }];
			}
			if (prefixLength >= token.end) {
				return [token];
			}
			return [];
		})
		.filter((token) => nextValue.slice(token.start, token.end) === token.label);
}

function normalizeSearchValue(value: string): string {
	return value.trim().toLowerCase();
}

export const StructuredComposer = forwardRef<StructuredComposerHandle, StructuredComposerProps>(
	function StructuredComposer(
		{ value, onChange, onSubmit, onPaste, onFocus, onBlur, placeholder, isProjectVariant },
		ref,
	) {
		const textareaRef = useRef<HTMLTextAreaElement>(null);
		const [tokens, setTokens] = useState<DirectiveToken[]>([]);
		const [trigger, setTrigger] = useState<ActiveTrigger | null>(null);
		const [activeIndex, setActiveIndex] = useState(0);
		const [scrollTop, setScrollTop] = useState(0);
		const composingRef = useRef(false);

		useEffect(() => {
			if (value) return;
			setTokens([]);
			setTrigger(null);
		}, [value]);

		const assistantOptions = useMemo<AssistantOption[]>(() => mockAssistants, []);

		const filteredAssistants = useMemo(() => {
			const query = normalizeSearchValue(trigger?.kind === "assistant" ? trigger.query : "");
			if (!query) return assistantOptions;
			return assistantOptions.filter((assistant) =>
				[assistant.name, assistant.code, assistant.description]
					.join(" ")
					.toLowerCase()
					.includes(query),
			);
		}, [assistantOptions, trigger]);

		const filteredCommands = useMemo(() => {
			const query = normalizeSearchValue(trigger?.kind === "command" ? trigger.query : "");
			if (!query) return mockChatCommands;
			return mockChatCommands.filter((command) =>
				[command.label, command.code, command.description, ...command.keywords]
					.join(" ")
					.toLowerCase()
					.includes(query),
			);
		}, [trigger]);

		const pickerItemCount =
			trigger?.kind === "assistant" ? filteredAssistants.length : filteredCommands.length;

		useEffect(() => {
			setActiveIndex(0);
		}, [trigger?.kind, trigger?.query]);

		const adjustHeight = useCallback(() => {
			const textarea = textareaRef.current;
			if (!textarea) return;
			textarea.style.height = "auto";
			textarea.style.height = `${Math.min(textarea.scrollHeight, 200)}px`;
		}, []);

		useEffect(() => {
			adjustHeight();
		}, [value, adjustHeight]);

		const validTokens = useMemo(
			() => tokens.filter((token) => value.slice(token.start, token.end) === token.label),
			[tokens, value],
		);

		const shouldUseHighlightLayer = validTokens.length > 0;

		const focusAt = useCallback((cursor: number) => {
			requestAnimationFrame(() => {
				const textarea = textareaRef.current;
				if (!textarea) return;
				textarea.focus();
				textarea.setSelectionRange(cursor, cursor);
			});
		}, []);

		const insertTrigger = useCallback(
			(kind: DirectiveKind) => {
				const textarea = textareaRef.current;
				const cursor = textarea?.selectionStart ?? value.length;
				const marker = kind === "assistant" ? "@" : "/";
				const needsLeadingSpace = cursor > 0 && !/\s/.test(value[cursor - 1] ?? "");
				const insertion = `${needsLeadingSpace ? " " : ""}${marker}`;
				const markerStart = cursor + (needsLeadingSpace ? 1 : 0);
				const nextValue = `${value.slice(0, cursor)}${insertion}${value.slice(cursor)}`;
				setTokens((current) => updateTokensForTextChange(current, value, nextValue));
				onChange(nextValue);
				setTrigger({ kind, start: markerStart, end: markerStart + 1, query: "" });
				focusAt(markerStart + 1);
			},
			[focusAt, onChange, value],
		);

		useImperativeHandle(
			ref,
			() => ({
				openAssistantPicker: () => insertTrigger("assistant"),
			}),
			[insertTrigger],
		);

		const selectToken = useCallback(
			(
				kind: DirectiveKind,
				option: AssistantOption | ChatCommand,
				activeTrigger: ActiveTrigger,
			) => {
				const isAssistant = kind === "assistant";
				const label = `${isAssistant ? "@" : "/"}${isAssistant ? (option as AssistantOption).name : (option as ChatCommand).label}`;
				const followingText = value.slice(activeTrigger.end);
				const trailingSpace = followingText.startsWith(" ") ? "" : " ";
				const nextValue = `${value.slice(0, activeTrigger.start)}${label}${trailingSpace}${followingText}`;
				const token: DirectiveToken = {
					id: `${kind}-${Date.now()}`,
					kind,
					start: activeTrigger.start,
					end: activeTrigger.start + label.length,
					label,
				};

				setTokens((current) => {
					const updated = updateTokensForTextChange(current, value, nextValue);
					return [...updated, token].sort((a, b) => a.start - b.start);
				});
				onChange(nextValue);
				setTrigger(null);
				focusAt(token.end + trailingSpace.length);
			},
			[focusAt, onChange, value],
		);

		const selectActiveItem = useCallback(() => {
			if (!trigger) return;
			if (trigger.kind === "assistant") {
				const assistant = filteredAssistants[activeIndex];
				if (assistant) selectToken("assistant", assistant, trigger);
				return;
			}
			const command = filteredCommands[activeIndex];
			if (command) selectToken("command", command, trigger);
		}, [activeIndex, filteredAssistants, filteredCommands, selectToken, trigger]);

		const handleChange = useCallback(
			(event: React.ChangeEvent<HTMLTextAreaElement>) => {
				const nextValue = event.target.value;
				const cursor = event.target.selectionStart ?? nextValue.length;
				setTokens((current) => updateTokensForTextChange(current, value, nextValue));
				onChange(nextValue);
				if (!composingRef.current) {
					setTrigger(findTrigger(nextValue, cursor));
				}
			},
			[onChange, value],
		);

		const handleKeyDown = useCallback(
			(event: React.KeyboardEvent<HTMLTextAreaElement>) => {
				if (trigger) {
					if (event.key === "ArrowDown" || event.key === "ArrowUp") {
						event.preventDefault();
						const direction = event.key === "ArrowDown" ? 1 : -1;
						setActiveIndex((current) => {
							if (pickerItemCount === 0) return 0;
							return (current + direction + pickerItemCount) % pickerItemCount;
						});
						return;
					}
					if ((event.key === "Enter" || event.key === "Tab") && pickerItemCount > 0) {
						event.preventDefault();
						selectActiveItem();
						return;
					}
					if (event.key === "Escape") {
						event.preventDefault();
						setTrigger(null);
						return;
					}
				}

				const submitByEnter = !isProjectVariant && event.key === "Enter" && !event.shiftKey;
				const submitByShortcut =
					isProjectVariant && event.key === "Enter" && (event.metaKey || event.ctrlKey);
				if (submitByEnter || submitByShortcut) {
					event.preventDefault();
					onSubmit();
				}
			},
			[isProjectVariant, onSubmit, pickerItemCount, selectActiveItem, trigger],
		);

		const renderHighlightedValue = () => {
			if (validTokens.length === 0) return value;

			const parts: React.ReactNode[] = [];
			let cursor = 0;
			for (const token of validTokens) {
				if (token.start < cursor) continue;
				parts.push(value.slice(cursor, token.start));
				parts.push(
					<span
						key={token.id}
						className={cn(
							"rounded-md ring-2",
							token.kind === "assistant"
								? "bg-blue-100 text-blue-700 ring-blue-100"
								: "bg-violet-100 text-violet-700 ring-violet-100",
						)}
					>
						{token.label}
					</span>,
				);
				cursor = token.end;
			}
			parts.push(value.slice(cursor));
			return parts;
		};

		const inputSpacingClass = isProjectVariant
			? "min-h-[92px] rounded-none px-0 py-0 text-base leading-7"
			: "min-h-[116px] rounded-2xl px-5 py-4 text-sm leading-6";

		return (
			<div className="relative">
				{trigger && (
					<div className="absolute bottom-full left-0 z-30 mb-2 w-full max-w-[360px] overflow-hidden rounded-2xl border border-slate-200/80 bg-white/95 p-1.5 shadow-[0_12px_36px_rgba(15,23,42,0.12)] backdrop-blur">
						<Command shouldFilter={false} className="rounded-xl! bg-transparent p-0">
							<div className="flex items-center gap-2 px-2.5 pb-1.5 pt-1 text-xs font-medium text-slate-400">
								{trigger.kind === "assistant" ? <>AI 队友</> : <>命令</>}
								{trigger.query && <span className="truncate text-slate-400">{trigger.query}</span>}
							</div>
							<CommandList className="max-h-60">
								<CommandEmpty className="py-8 text-slate-400">没有匹配项</CommandEmpty>
								<CommandGroup className="p-0">
									{trigger.kind === "assistant"
										? filteredAssistants.map((assistant, index) => (
												<CommandItem
													key={assistant.code}
													value={assistant.code}
													onMouseDown={(event) => event.preventDefault()}
													onSelect={() => selectToken("assistant", assistant, trigger)}
													className={cn(
														"rounded-xl px-2.5 py-2",
														index === activeIndex && "bg-slate-100",
													)}
												>
													<div className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
														<Bot className="size-4" />
													</div>
													<div className="min-w-0 flex-1">
														<div className="truncate font-medium text-slate-700">
															{assistant.name}
														</div>
														<div className="truncate text-xs text-slate-400">
															{assistant.description}
														</div>
													</div>
												</CommandItem>
											))
										: filteredCommands.map((command, index) => (
												<CommandItem
													key={command.code}
													value={command.code}
													onMouseDown={(event) => event.preventDefault()}
													onSelect={() => selectToken("command", command, trigger)}
													className={cn(
														"rounded-xl px-2.5 py-2",
														index === activeIndex && "bg-slate-100",
													)}
												>
													{/* <div className="flex size-7 shrink-0 items-center justify-center rounded-md bg-violet-50 text-violet-600">
														<Slash className="size-4" />
													</div> */}
													<div className="min-w-0 flex-1">
														<div className="font-medium">/{command.label}</div>
														<div className="truncate text-xs text-slate-400">
															{command.description}
														</div>
													</div>
												</CommandItem>
											))}
								</CommandGroup>
							</CommandList>
						</Command>
					</div>
				)}

				{shouldUseHighlightLayer && (
					<div
						aria-hidden="true"
						className={cn(
							"pointer-events-none absolute inset-0 overflow-hidden whitespace-pre-wrap break-words text-slate-700",
							inputSpacingClass,
						)}
					>
						{/* 只有真正需要高亮 token 时才启用镜像层，避免长文本粘贴时双层排版产生重影。 */}
						<div style={{ transform: `translateY(-${scrollTop}px)` }}>
							{renderHighlightedValue()}
						</div>
					</div>
				)}
				<textarea
					ref={textareaRef}
					value={value}
					onChange={handleChange}
					onKeyDown={handleKeyDown}
					onPaste={onPaste}
					onFocus={onFocus}
					onBlur={() => {
						onBlur();
						setTimeout(() => setTrigger(null), 100);
					}}
					onScroll={(event) => setScrollTop(event.currentTarget.scrollTop)}
					onCompositionStart={() => {
						composingRef.current = true;
					}}
					onCompositionEnd={(event) => {
						composingRef.current = false;
						const cursor = event.currentTarget.selectionStart ?? value.length;
						setTrigger(findTrigger(event.currentTarget.value, cursor));
					}}
					placeholder={placeholder}
					className={cn(
						"relative z-10 max-h-[220px] w-full resize-none bg-transparent caret-slate-700 focus:outline-none placeholder:text-slate-400",
						shouldUseHighlightLayer ? "text-transparent" : "text-slate-700",
						inputSpacingClass,
					)}
					rows={1}
				/>
			</div>
		);
	},
);
