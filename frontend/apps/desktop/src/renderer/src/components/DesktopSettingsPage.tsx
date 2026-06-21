import { useEffect, useState, type ReactNode } from "react";
import { ArrowUpCircle, CheckCircle2, Download, LoaderCircle, RefreshCcw, RotateCw } from "lucide-react";
import { toast } from "sonner";
import { Badge } from "@leros/ui/components/ui/badge";
import { Button } from "@leros/ui/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@leros/ui/components/ui/card";
import { Progress } from "@leros/ui/components/ui/progress";
import { Separator } from "@leros/ui/components/ui/separator";
import type { DesktopUpdateState } from "../../../shared/auto-update";

const initialState: DesktopUpdateState = {
	currentVersion: "0.0.0",
	phase: "idle",
	message: "正在读取版本信息",
	canCheck: false,
	canRestart: false,
};

const phaseLabelMap: Record<DesktopUpdateState["phase"], string> = {
	idle: "待检查",
	checking: "检查中",
	available: "已发现",
	downloading: "下载中",
	downloaded: "可安装",
	"up-to-date": "最新",
	error: "异常",
	unsupported: "不可用",
};

function formatTime(value?: string) {
	if (!value) {
		return "未检查";
	}

	return new Intl.DateTimeFormat("zh-CN", {
		dateStyle: "medium",
		timeStyle: "short",
	}).format(new Date(value));
}

function getPhaseVariant(phase: DesktopUpdateState["phase"]): "default" | "secondary" | "destructive" | "outline" {
	if (phase === "downloaded") {
		return "default";
	}

	if (phase === "error") {
		return "destructive";
	}

	if (phase === "available" || phase === "downloading" || phase === "checking") {
		return "secondary";
	}

	return "outline";
}

export function DesktopSettingsPage() {
	const [updateState, setUpdateState] = useState<DesktopUpdateState>(initialState);
	const [checking, setChecking] = useState(false);
	const [restarting, setRestarting] = useState(false);

	useEffect(() => {
		let mounted = true;

		void window.lerosDesktop.getState().then((state) => {
			if (mounted) {
				setUpdateState(state);
			}
		});

		const unsubscribe = window.lerosDesktop.subscribe((state) => {
			setUpdateState(state);
		});

		return () => {
			mounted = false;
			unsubscribe();
		};
	}, []);

	const handleCheckForUpdates = async () => {
		setChecking(true);
		try {
			const nextState = await window.lerosDesktop.checkForUpdates();
			setUpdateState(nextState);
			if (nextState.phase === "unsupported") {
				toast.message(nextState.message);
			}
			if (nextState.phase === "up-to-date") {
				toast.success("当前已经是最新版本");
			}
		} finally {
			setChecking(false);
		}
	};

	const handleRestartToUpdate = async () => {
		setRestarting(true);
		try {
			const accepted = await window.lerosDesktop.quitAndInstall();
			if (!accepted) {
				toast.message("当前还没有可安装的更新");
			}
		} finally {
			setRestarting(false);
		}
	};

	return (
		<div className="min-h-0 flex-1 overflow-y-auto bg-[linear-gradient(180deg,#f4f7fb_0%,#eef3ff_100%)]">
			<div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-6 py-8">
				<div className="space-y-2">
					<h1 className="text-2xl font-semibold tracking-tight text-slate-950">桌面端更新</h1>
					<p className="max-w-2xl text-sm leading-6 text-slate-600">
						应用会在启动后自动检查更新，并从对象存储 + CDN 的静态发布目录下载新版本。
					</p>
				</div>

				<Card className="border-white/70 bg-white/90 shadow-sm backdrop-blur">
					<CardHeader>
						<CardTitle>当前版本</CardTitle>
						<CardDescription>已安装用户会从稳定版更新目录拉取最新安装包元数据。</CardDescription>
					</CardHeader>
					<CardContent className="space-y-5">
						<div className="flex flex-wrap items-center gap-3">
							<div className="text-3xl font-semibold tracking-tight text-slate-950">
								v{updateState.currentVersion}
							</div>
							<Badge variant={getPhaseVariant(updateState.phase)}>{phaseLabelMap[updateState.phase]}</Badge>
						</div>

						<div className="grid gap-4 md:grid-cols-2">
							<StatusItem label="最近检查" value={formatTime(updateState.lastCheckedAt)} />
							<StatusItem label="状态说明" value={updateState.message} />
							<StatusItem
								label="可用版本"
								value={updateState.availableVersion ? `v${updateState.availableVersion}` : "暂无"}
							/>
							<StatusItem
								label="已下载版本"
								value={updateState.downloadedVersion ? `v${updateState.downloadedVersion}` : "暂无"}
							/>
						</div>

						{typeof updateState.progressPercent === "number" ? (
							<div className="space-y-2">
								<div className="flex items-center justify-between text-sm text-slate-600">
									<span>下载进度</span>
									<span>{Math.round(updateState.progressPercent)}%</span>
								</div>
								<Progress value={updateState.progressPercent} />
							</div>
						) : null}

						{updateState.releaseNotes ? (
							<>
								<Separator />
								<div className="space-y-2">
									<div className="text-sm font-medium text-slate-900">更新说明</div>
									<pre className="whitespace-pre-wrap rounded-xl bg-slate-950/[0.03] p-4 text-sm leading-6 text-slate-700">
										{updateState.releaseNotes}
									</pre>
								</div>
							</>
						) : null}
					</CardContent>
					<CardFooter className="flex flex-wrap justify-between gap-3">
						<div className="text-xs leading-5 text-slate-500">
							macOS 使用 `latest-mac.yml`，Windows 使用 `latest.yml`，文件由发布流水线上传到 CDN。
						</div>
						<div className="flex flex-wrap items-center gap-2">
							<Button
								type="button"
								variant="outline"
								onClick={handleCheckForUpdates}
								disabled={!updateState.canCheck || checking}
							>
								{checking || updateState.phase === "checking" ? (
									<LoaderCircle className="animate-spin" />
								) : (
									<RefreshCcw />
								)}
								检查更新
							</Button>
							<Button
								type="button"
								onClick={handleRestartToUpdate}
								disabled={!updateState.canRestart || restarting}
							>
								{restarting ? <RotateCw className="animate-spin" /> : <ArrowUpCircle />}
								重启并安装
							</Button>
						</div>
					</CardFooter>
				</Card>

				<Card className="border-white/70 bg-white/80 shadow-sm backdrop-blur">
					<CardHeader>
						<CardTitle>发布目录约定</CardTitle>
						<CardDescription>第一版使用稳定版静态目录即可，无需额外更新服务。</CardDescription>
					</CardHeader>
					<CardContent className="grid gap-4 md:grid-cols-3">
						<GuideItem
							icon={<CheckCircle2 className="size-4" />}
							title="对象存储"
							description="上传安装包、blockmap 和 latest 元数据文件。"
						/>
						<GuideItem
							icon={<Download className="size-4" />}
							title="CDN 分发"
							description="将更新目录暴露为 HTTPS 静态地址，例如 /desktop/stable。"
						/>
						<GuideItem
							icon={<ArrowUpCircle className="size-4" />}
							title="客户端拉取"
							description="启动后自动检查，下载完成后由用户确认重启安装。"
						/>
					</CardContent>
				</Card>
			</div>
		</div>
	);
}

function StatusItem({ label, value }: { label: string; value: string }) {
	return (
		<div className="rounded-2xl border border-slate-200/80 bg-slate-50/70 p-4">
			<div className="text-xs font-medium tracking-wide text-slate-500 uppercase">{label}</div>
			<div className="mt-2 text-sm leading-6 text-slate-800">{value}</div>
		</div>
	);
}

function GuideItem({
	icon,
	title,
	description,
}: {
	icon: ReactNode;
	title: string;
	description: string;
}) {
	return (
		<div className="rounded-2xl border border-slate-200/80 bg-slate-50/70 p-4">
			<div className="flex items-center gap-2 text-sm font-medium text-slate-900">
				<span className="flex size-7 items-center justify-center rounded-full bg-white text-slate-700 shadow-sm">
					{icon}
				</span>
				{title}
			</div>
			<p className="mt-3 text-sm leading-6 text-slate-600">{description}</p>
		</div>
	);
}
