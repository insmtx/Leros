import { useEffect, useRef, useState } from "react";
import { ThemeProvider } from "@leros/ui/components/common/theme-provider";
import { Button } from "@leros/ui/components/ui/button";
import { Toaster } from "@leros/ui/components/ui/sonner";
import { ArrowUpCircle, X } from "lucide-react";
import { HashRouter } from "react-router-dom";
import { AppRoutes } from "./routes";
import type { DesktopUpdateState } from "../../shared/auto-update";

const updatePromptSnoozeMs = 5 * 60 * 1000;

export default function App() {
	return (
		<HashRouter>
			<ThemeProvider defaultTheme="system">
				<DesktopUpdateNotifier />
				<AppRoutes />
				<Toaster />
			</ThemeProvider>
		</HashRouter>
	);
}

function DesktopUpdateNotifier() {
	const previousPhaseRef = useRef<DesktopUpdateState["phase"] | null>(null);
	const previousVersionRef = useRef<string | undefined>(undefined);
	const snoozeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const [promptOpen, setPromptOpen] = useState(false);
	const [downloadedVersion, setDownloadedVersion] = useState<string | undefined>(undefined);
	const [installing, setInstalling] = useState(false);
	const [installError, setInstallError] = useState<string | null>(null);

	const clearSnoozeTimer = () => {
		if (snoozeTimerRef.current) {
			clearTimeout(snoozeTimerRef.current);
			snoozeTimerRef.current = null;
		}
	};

	const openUpdatePrompt = (version?: string) => {
		clearSnoozeTimer();
		setDownloadedVersion(version);
		setInstallError(null);
		setPromptOpen(true);
	};

	const snoozeUpdatePrompt = () => {
		setPromptOpen(false);

		if (!downloadedVersion || installing) {
			return;
		}

		clearSnoozeTimer();
		snoozeTimerRef.current = setTimeout(() => {
			setPromptOpen(true);
			snoozeTimerRef.current = null;
		}, updatePromptSnoozeMs);
	};

	useEffect(() => {
		let mounted = true;

		void window.lerosDesktop.getState().then((state) => {
			if (!mounted) {
				return;
			}

			previousPhaseRef.current = state.phase;
			previousVersionRef.current = state.availableVersion ?? state.downloadedVersion;

			if (state.phase === "downloaded") {
				openUpdatePrompt(state.downloadedVersion ?? state.availableVersion);
			}
		});

		const unsubscribe = window.lerosDesktop.subscribe((state) => {
			const previousPhase = previousPhaseRef.current;
			const previousVersion = previousVersionRef.current;
			const nextVersion = state.availableVersion ?? state.downloadedVersion;

			if (
				state.phase === "downloaded" &&
				(previousPhase !== "downloaded" || previousVersion !== nextVersion)
			) {
				openUpdatePrompt(nextVersion);
			}

			previousPhaseRef.current = state.phase;
			previousVersionRef.current = nextVersion;
		});

		return () => {
			mounted = false;
			clearSnoozeTimer();
			unsubscribe();
		};
	}, []);

	const handleInstallNow = async () => {
		setInstalling(true);
		setInstallError(null);
		try {
			const accepted = await window.lerosDesktop.quitAndInstall();
			if (!accepted) {
				setInstalling(false);
				setInstallError("当前还没有可安装的更新");
			}
		} catch (error) {
			setInstalling(false);
			setInstallError(error instanceof Error ? error.message : "启动安装失败，请稍后重试");
		}
	};

	if (!promptOpen) {
		return null;
	}

	return (
		<div className="pointer-events-none fixed right-2 bottom-4 z-50 flex max-w-[calc(100vw-1rem)] justify-end">
			<div className="pointer-events-auto relative w-auto max-w-full rounded-2xl border border-slate-200/90 bg-white px-3.5 py-2 text-slate-900 shadow-[0_16px_48px_rgba(15,23,42,0.14)]">
				<Button
					type="button"
					variant="ghost"
					size="icon-xs"
					className="absolute -top-2 -right-2 rounded-full border border-slate-200 bg-white text-slate-500 shadow-sm hover:bg-white hover:text-slate-900"
					onClick={snoozeUpdatePrompt}
					disabled={installing}
					aria-label="稍后安装"
				>
					<X className="size-3" />
				</Button>
				<div className="flex max-w-full items-center gap-2.5">
					<div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-emerald-50 text-emerald-500 ring-1 ring-emerald-100">
						<ArrowUpCircle className="size-3.5" />
					</div>
					<div className="min-w-[170px] max-w-[200px] flex-[0_1_auto]">
						<div className="truncate text-[13px] font-semibold leading-4">新版本已经下载完成</div>
						<div className="truncate text-[11px] leading-4 text-slate-500">
							{downloadedVersion ? `v${downloadedVersion} 已准备就绪。` : "更新包已准备就绪。"}
						</div>
					</div>
					<div className="flex shrink-0 items-center gap-2">
						<Button
							type="button"
							variant="outline"
							size="sm"
							className="h-7 min-w-14 rounded-xl border-slate-200 bg-white px-3 text-xs text-slate-700 hover:bg-slate-50"
							onClick={snoozeUpdatePrompt}
							disabled={installing}
						>
							稍后
						</Button>
						<Button
							type="button"
							size="sm"
							className="h-7 min-w-20 rounded-xl bg-slate-950 px-3 text-xs text-white hover:bg-slate-800"
							onClick={handleInstallNow}
							disabled={installing}
						>
							{installing ? "正在启动..." : "重启升级"}
						</Button>
					</div>
				</div>
				{installError ? <div className="mt-2 truncate text-xs text-red-500">{installError}</div> : null}
			</div>
		</div>
	);
}
