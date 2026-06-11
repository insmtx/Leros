export function formatTime(timestamp: number): string {
	const date = new Date(timestamp);
	return date.toLocaleTimeString("zh-CN", {
		hour: "2-digit",
		minute: "2-digit",
	});
}

export function formatDate(timestamp: number): string {
	const date = new Date(timestamp);
	const isToday = date.toDateString() === new Date().toDateString();
	if (isToday) {
		return `今天 ${formatTime(timestamp)}`;
	}
	return date.toLocaleDateString("zh-CN", {
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit",
	});
}

export function formatFileSize(bytes: number): string {
	if (bytes < 1024) return `${bytes}B`;
	if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
	return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
}

export function formatTokenCount(count: number): string {
	if (!count) return "0";
	if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
	if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
	return String(count);
}
