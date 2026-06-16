"use client";

import { AlertCircle, Table2 } from "lucide-react";
import { useMemo, useState } from "react";
import { read, utils, type WorkBook } from "xlsx";

const MAX_PREVIEW_ROWS = 200;
const MAX_PREVIEW_COLUMNS = 50;

type SheetPreview = {
	name: string;
	rows: string[][];
	totalRows: number;
	totalColumns: number;
};

export function SpreadsheetPreview({
	buffer,
	fileName,
}: {
	buffer: ArrayBuffer;
	fileName: string;
}) {
	const result = useMemo(() => parseWorkbook(buffer), [buffer]);
	const [selectedSheet, setSelectedSheet] = useState(0);

	if (result.error) {
		return (
			<div className="flex h-full min-h-[320px] items-center justify-center px-8 text-center text-sm text-[var(--leros-text-muted)]">
				<div>
					<AlertCircle className="mx-auto mb-3 size-8 text-[var(--leros-danger)]" />
					<p>无法解析表格文件</p>
					<p className="mt-1 text-xs">{result.error}</p>
				</div>
			</div>
		);
	}

	const activeSheetIndex = result.sheets[selectedSheet] ? selectedSheet : 0;
	const sheet = result.sheets[activeSheetIndex];
	if (!sheet) {
		return (
			<div className="flex h-full min-h-[320px] items-center justify-center text-sm text-[var(--leros-text-muted)]">
				文件中没有可预览的工作表
			</div>
		);
	}

	const truncated = sheet.totalRows > MAX_PREVIEW_ROWS || sheet.totalColumns > MAX_PREVIEW_COLUMNS;
	const visibleColumns = Math.min(sheet.totalColumns, MAX_PREVIEW_COLUMNS);

	return (
		<div className="flex h-full min-h-[320px] flex-col overflow-hidden bg-white">
			<div className="flex shrink-0 items-center gap-2 border-b border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] px-3 py-2">
				<Table2 className="size-4 shrink-0 text-emerald-600" />
				<div className="min-w-0 flex-1 truncate text-xs text-[var(--leros-text-muted)]">
					{fileName} · {sheet.totalRows.toLocaleString()} 行 × {sheet.totalColumns.toLocaleString()}{" "}
					列
				</div>
				{truncated && (
					<div className="shrink-0 text-xs text-amber-700">
						仅展示前 {MAX_PREVIEW_ROWS} 行、{MAX_PREVIEW_COLUMNS} 列
					</div>
				)}
			</div>

			<div className="min-h-0 flex-1 overflow-auto">
				<table className="border-separate border-spacing-0 text-xs text-[var(--leros-text)]">
					<thead>
						<tr>
							<th className="sticky left-0 top-0 z-30 min-w-12 border-b border-r border-[var(--leros-control-border)] bg-slate-200" />
							{Array.from({ length: visibleColumns }, (_, columnIndex) => (
								<th
									key={columnIndex}
									className="sticky top-0 z-10 min-w-28 border-b border-r border-[var(--leros-control-border)] bg-slate-100 px-2.5 py-1.5 text-center font-medium text-slate-500"
								>
									{columnName(columnIndex)}
								</th>
							))}
						</tr>
					</thead>
					<tbody>
						{sheet.rows.map((row, rowIndex) => (
							<tr key={`${sheet.name}-${rowIndex}`}>
								<th className="sticky left-0 z-20 min-w-12 border-b border-r border-[var(--leros-control-border)] bg-slate-100 px-2 py-1.5 text-center font-normal text-slate-500">
									{rowIndex + 1}
								</th>
								{row.map((cell, columnIndex) => (
									<td
										key={`${sheet.name}-${rowIndex}-${columnIndex}`}
										title={cell}
										className="max-w-64 min-w-28 truncate whitespace-nowrap border-b border-r border-[var(--leros-control-border)] bg-white px-2.5 py-1.5"
									>
										{cell}
									</td>
								))}
							</tr>
						))}
					</tbody>
				</table>
			</div>

			{result.sheets.length > 1 && (
				<div className="flex shrink-0 gap-1 overflow-x-auto border-t border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] px-2 py-1.5">
					{result.sheets.map((item, index) => (
						<button
							key={item.name}
							type="button"
							onClick={() => setSelectedSheet(index)}
							className={`shrink-0 rounded-md px-3 py-1.5 text-xs transition-colors ${
								index === activeSheetIndex
									? "bg-white font-medium text-emerald-700 shadow-sm ring-1 ring-[var(--leros-control-border)]"
									: "text-[var(--leros-text-muted)] hover:bg-white/70"
							}`}
						>
							{item.name}
						</button>
					))}
				</div>
			)}
		</div>
	);
}

function columnName(columnIndex: number): string {
	let value = columnIndex + 1;
	let name = "";
	while (value > 0) {
		value -= 1;
		name = String.fromCharCode(65 + (value % 26)) + name;
		value = Math.floor(value / 26);
	}
	return name;
}

function parseWorkbook(buffer: ArrayBuffer):
	| { sheets: SheetPreview[]; error?: undefined }
	| {
			sheets: [];
			error: string;
	  } {
	try {
		const workbook = read(buffer, { type: "array", cellDates: true });
		return { sheets: workbook.SheetNames.map((name) => parseSheet(workbook, name)) };
	} catch (err) {
		return {
			sheets: [],
			error: err instanceof Error ? err.message : "表格格式不正确或文件已损坏",
		};
	}
}

function parseSheet(workbook: WorkBook, name: string): SheetPreview {
	const worksheet = workbook.Sheets[name];
	if (!worksheet?.["!ref"]) {
		return { name, rows: [], totalRows: 0, totalColumns: 0 };
	}

	const range = utils.decode_range(worksheet["!ref"]);
	const totalRows = range.e.r - range.s.r + 1;
	const totalColumns = range.e.c - range.s.c + 1;
	const rows = utils.sheet_to_json<string[]>(worksheet, {
		header: 1,
		raw: false,
		defval: "",
		range: {
			s: range.s,
			e: {
				r: Math.min(range.e.r, range.s.r + MAX_PREVIEW_ROWS - 1),
				c: Math.min(range.e.c, range.s.c + MAX_PREVIEW_COLUMNS - 1),
			},
		},
	});

	return {
		name,
		rows: rows.map((row) => row.map((cell) => String(cell))),
		totalRows,
		totalColumns,
	};
}
