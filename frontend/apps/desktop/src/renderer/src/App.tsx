import { useEffect, useRef, useState } from "react";
import { ThemeProvider } from "@leros/ui/components/common/theme-provider";
import { Button } from "@leros/ui/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@leros/ui/components/ui/dialog";
import { Toaster } from "@leros/ui/components/ui/sonner";
import { toast } from "sonner";
import { HashRouter } from "react-router-dom";
import { AppRoutes } from "./routes";
import type { DesktopUpdateState } from "../../shared/auto-update";

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
	const [confirmOpen, setConfirmOpen] = useState(false);
	const [downloadedVersion, setDownloadedVersion] = useState<string | undefined>(undefined);
	const [installing, setInstalling] = useState(false);

	useEffect(() => {
		let mounted = true;

		void window.lerosDesktop.getState().then((state) => {
			if (!mounted) {
				return;
			}

			previousPhaseRef.current = state.phase;
			previousVersionRef.current = state.availableVersion ?? state.downloadedVersion;
		});

		const unsubscribe = window.lerosDesktop.subscribe((state) => {
			const previousPhase = previousPhaseRef.current;
			const previousVersion = previousVersionRef.current;
			const nextVersion = state.availableVersion ?? state.downloadedVersion;

			if (
				state.phase === "available" &&
				(previousPhase !== "available" || previousVersion !== nextVersion)
			) {
				toast.message(`发现新版本 ${nextVersion ? `v${nextVersion}` : ""}`.trim(), {
					description: "正在后台下载更新包",
				});
			}

			if (
				state.phase === "downloaded" &&
				(previousPhase !== "downloaded" || previousVersion !== nextVersion)
			) {
				setDownloadedVersion(nextVersion);
				setConfirmOpen(true);
				toast.success("新版本已经下载完成", {
					description: "确认后可立即重启并安装更新",
				});
			}

			if (state.phase === "error" && previousPhase !== "error") {
				toast.error("更新失败", {
					description: state.message,
				});
			}

			previousPhaseRef.current = state.phase;
			previousVersionRef.current = nextVersion;
		});

		return () => {
			mounted = false;
			unsubscribe();
		};
	}, []);

	const handleInstallNow = async () => {
		setInstalling(true);
		try {
			const accepted = await window.lerosDesktop.quitAndInstall();
			if (!accepted) {
				toast.message("当前还没有可安装的更新");
				setInstalling(false);
				setConfirmOpen(false);
			}
		} catch (error) {
			setInstalling(false);
			toast.error("启动安装失败", {
				description: error instanceof Error ? error.message : "请稍后重试",
			});
		}
	};

	return (
		<Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
			<DialogContent className="sm:max-w-md" showCloseButton={!installing}>
				<DialogHeader>
					<DialogTitle>新版本已经下载完成</DialogTitle>
					<DialogDescription>
						{downloadedVersion ? `v${downloadedVersion} 已准备就绪。` : "更新包已准备就绪。"}
						是否立即重启应用并安装更新？
					</DialogDescription>
				</DialogHeader>
				<DialogFooter className="mt-4">
					<Button variant="outline" onClick={() => setConfirmOpen(false)} disabled={installing}>
						稍后
					</Button>
					<Button onClick={handleInstallNow} disabled={installing}>
						{installing ? "正在启动安装…" : "立即重启安装"}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
