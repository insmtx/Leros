"use client";

import { Button } from "@leros/ui/components/ui/button";
import { cn } from "@leros/ui/lib/utils";
import {
	BarChart3,
	FileText,
	Globe,
	Languages,
	Palette,
	Plus,
	Search,
	SlidersHorizontal,
	Star,
	Terminal,
	TrendingUp,
} from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

// ─── Mock data ───────────────────────────────────────────────────────────────

interface SkillItem {
	id: string;
	name: string;
	verified: boolean;
	publisher: string;
	category: string;
	tags: string[];
	description: string;
	rating: number;
	installed: boolean;
	icon: string;
}

const MOCK_SKILLS: SkillItem[] = [
	{ id: "1", name: "高级 SQL 查询生成器", verified: true, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["Data", "SQL"], description: "将复杂的自然语言问题转化为经过优化的、特定数据库方言的 SQL 查询。支持 PostgreSQL、Snowflake 和 BigQuery。", rating: 4.9, installed: false, icon: "query_stats" },
	{ id: "2", name: "法律文档摘要提取", verified: false, publisher: "由 Leros AI 提供", category: "自然语言", tags: ["NLP", "Legal"], description: "从冗长的法律合同中提取关键条款、义务和责任。生成优先考虑风险评估和合规性标记的结构化摘要。", rating: 4.7, installed: false, icon: "description" },
	{ id: "3", name: "REST API 脚手架引擎", verified: true, publisher: "由 Leros AI 提供", category: "代码生成", tags: ["Code", "API"], description: "从自然语言需求或 OpenAPI 规范自动生成完整的、生产就绪的 RESTful API 端点。包含基础测试。", rating: 5.0, installed: true, icon: "terminal" },
	{ id: "4", name: "极简 UI 生成助手", verified: false, publisher: "由 Leros AI 提供", category: "视觉/媒体", tags: ["Design", "Tailwind"], description: "根据功能描述自动推荐并生成符合建筑极简主义风格的 UI 组件库。支持 Tailwind CSS 输出。", rating: 4.5, installed: false, icon: "palette" },
	{ id: "5", name: "高保真多语言翻译器", verified: false, publisher: "由 Leros AI 提供", category: "自然语言", tags: ["Language"], description: "基于大规模多语言模型的翻译工具，能够完美保留原文的语气、语境和行业术语。", rating: 4.8, installed: false, icon: "translate" },
	{ id: "6", name: "市场洞察全能助手", verified: false, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["Market", "Analytics"], description: "实时抓取全网市场数据，自动生成竞争对手分析报告和行业趋势预测图表。", rating: 4.6, installed: false, icon: "insights" },
	{ id: "7", name: "智能代码审查员", verified: true, publisher: "由 Leros AI 提供", category: "代码生成", tags: ["Code", "Review"], description: "自动审查 PR 代码，检测潜在 bug、安全漏洞和性能瓶颈，提供可操作的修复建议。", rating: 4.8, installed: false, icon: "terminal" },
	{ id: "8", name: "电商数据分析看板", verified: false, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["E-commerce", "BI"], description: "连接电商平台 API，实时生成销量、库存和用户行为分析可视化看板。", rating: 4.4, installed: false, icon: "query_stats" },
	{ id: "9", name: "视频脚本自动生成", verified: true, publisher: "由 Leros AI 提供", category: "视觉/媒体", tags: ["Video", "Content"], description: "根据选题或关键词自动生成视频脚本、分镜描述和旁白文案。", rating: 4.6, installed: false, icon: "palette" },
	{ id: "10", name: "多语种合同翻译", verified: false, publisher: "由 Leros AI 提供", category: "自然语言", tags: ["Legal", "Translate"], description: "精准翻译多语种商业合同，自动识别法律术语并保持条款一致性。", rating: 4.9, installed: false, icon: "description" },
	{ id: "11", name: "智能日志分析引擎", verified: true, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["DevOps", "Log"], description: "实时聚合和分析应用日志，自动检测异常模式并触发告警通知。", rating: 4.7, installed: false, icon: "insights" },
	{ id: "12", name: "PPT 智能排版助手", verified: false, publisher: "由 Leros AI 提供", category: "视觉/媒体", tags: ["PPT", "Design"], description: "将文字大纲一键转化为排版精美的演示文稿，支持多种风格模板。", rating: 4.3, installed: false, icon: "palette" },
];

const CATEGORIES = [
	{ value: "", label: "全部" },
	{ value: "数据分析", label: "数据分析" },
	{ value: "自然语言", label: "自然语言" },
	{ value: "视觉/媒体", label: "视觉/媒体" },
	{ value: "代码生成", label: "代码生成" },
];

// ─── Icon mapping ────────────────────────────────────────────────────────────

const iconMap: Record<string, React.ReactNode> = {
	query_stats: <BarChart3 className="size-[18px]" />,
	translate: <Languages className="size-[18px]" />,
	terminal: <Terminal className="size-[18px]" />,
	palette: <Palette className="size-[18px]" />,
	description: <FileText className="size-[18px]" />,
	insights: <TrendingUp className="size-[18px]" />,
};

// Extra pool for infinite-scroll loading (2 batches × 12)
const MORE_SKILL_TEMPLATES: SkillItem[] = [
	{ id: "t1", name: "单元测试自动生成", verified: true, publisher: "由 Leros AI 提供", category: "代码生成", tags: ["Test", "CI"], description: "扫描源代码自动生成高覆盖率的单元测试用例，支持 Jest、pytest 等框架。", rating: 4.8, installed: false, icon: "terminal" },
	{ id: "t2", name: "舆情情感分析", verified: false, publisher: "由 Leros AI 提供", category: "自然语言", tags: ["Sentiment", "Social"], description: "实时抓取社交媒体讨论，分析公众对品牌/事件的情感倾向和趋势变化。", rating: 4.5, installed: false, icon: "translate" },
	{ id: "t3", name: "智能发票识别", verified: true, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["OCR", "Finance"], description: "自动识别和提取发票、收据中的关键字段，支持批量导出至财务系统。", rating: 4.7, installed: false, icon: "query_stats" },
	{ id: "t4", name: "代码性能剖析", verified: false, publisher: "由 Leros AI 提供", category: "代码生成", tags: ["Perf", "Profile"], description: "对代码进行深度性能分析，定位热点函数并提供优化方案。支持多种编程语言。", rating: 4.6, installed: false, icon: "terminal" },
	{ id: "t5", name: "智能海报设计", verified: true, publisher: "由 Leros AI 提供", category: "视觉/媒体", tags: ["Design", "Marketing"], description: "输入营销文案自动生成多尺寸社交媒体海报，支持品牌色和 VI 规范配置。", rating: 4.4, installed: false, icon: "palette" },
	{ id: "t6", name: "会议纪要生成器", verified: false, publisher: "由 Leros AI 提供", category: "自然语言", tags: ["Meeting", "Summary"], description: "导入会议录音或文字记录，自动生成结构化的会议纪要和待办事项清单。", rating: 4.9, installed: false, icon: "description" },
	{ id: "t7", name: "库存预测引擎", verified: true, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["Forecast", "Supply"], description: "基于历史销售数据和季节性因素，利用 ML 模型预测未来库存需求。", rating: 4.5, installed: false, icon: "insights" },
	{ id: "t8", name: "API 文档生成器", verified: false, publisher: "由 Leros AI 提供", category: "代码生成", tags: ["Docs", "API"], description: "扫描源码自动生成 OpenAPI/Swagger 文档，保持与代码同步更新。", rating: 4.7, installed: false, icon: "terminal" },
	{ id: "t9", name: "短视频剪辑助手", verified: true, publisher: "由 Leros AI 提供", category: "视觉/媒体", tags: ["Video", "Edit"], description: "根据文字脚本自动剪辑短视频，智能匹配转场效果和背景音乐。", rating: 4.3, installed: false, icon: "palette" },
	{ id: "t10", name: "学术论文润色", verified: false, publisher: "由 Leros AI 提供", category: "自然语言", tags: ["Academic", "Writing"], description: "对英文学术论文进行语法、用词和逻辑结构润色，支持多种引用格式。", rating: 4.8, installed: false, icon: "translate" },
	{ id: "t11", name: "用户画像构建", verified: true, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["Profile", "CRM"], description: "整合多源用户行为数据，自动构建 360° 用户画像和分群标签体系。", rating: 4.6, installed: false, icon: "query_stats" },
	{ id: "t12", name: "微服务拆分顾问", verified: false, publisher: "由 Leros AI 提供", category: "代码生成", tags: ["Architecture", "Micro"], description: "分析单体应用代码结构，推荐微服务拆分方案并生成服务边界定义。", rating: 4.4, installed: false, icon: "terminal" },
	// batch 2
	{ id: "t13", name: "图标自动生成器", verified: true, publisher: "由 Leros AI 提供", category: "视觉/媒体", tags: ["Icon", "SVG"], description: "通过自然语言描述生成自定义 SVG 图标，支持多种风格和导出格式。", rating: 4.7, installed: false, icon: "palette" },
	{ id: "t14", name: "客户邮件分类", verified: false, publisher: "由 Leros AI 提供", category: "自然语言", tags: ["Email", "Classify"], description: "自动分类客户邮件意图，支持自定义标签和自动路由到对应处理队列。", rating: 4.6, installed: false, icon: "description" },
	{ id: "t15", name: "A/B 实验分析", verified: true, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["A/B", "Stats"], description: "对 A/B 实验数据进行统计分析，输出置信区间、显著性水平和决策建议。", rating: 4.8, installed: false, icon: "insights" },
	{ id: "t16", name: "数据库迁移脚本", verified: false, publisher: "由 Leros AI 提供", category: "代码生成", tags: ["DB", "Migrate"], description: "自动生成安全的数据迁移脚本，支持 schema 变更检测和回滚方案。", rating: 4.5, installed: false, icon: "terminal" },
	{ id: "t17", name: "品牌 VI 检测", verified: true, publisher: "由 Leros AI 提供", category: "视觉/媒体", tags: ["Brand", "Check"], description: "上传设计稿自动检测是否符合品牌 VI 规范，输出逐项评分和修正建议。", rating: 4.3, installed: false, icon: "palette" },
	{ id: "t18", name: "对话意图识别", verified: false, publisher: "由 Leros AI 提供", category: "自然语言", tags: ["Intent", "Chat"], description: "实时识别客户对话中的意图和情绪，支持多轮上下文理解与槽位填充。", rating: 4.9, installed: false, icon: "translate" },
	{ id: "t19", name: "实时数据大屏", verified: true, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["Dashboard", "Live"], description: "连接实时数据源生成可视化大屏，支持拖拽配置图表布局和自动刷新。", rating: 4.7, installed: false, icon: "query_stats" },
	{ id: "t20", name: "正则表达式生成", verified: false, publisher: "由 Leros AI 提供", category: "代码生成", tags: ["Regex", "Tool"], description: "用自然语言描述匹配规则，自动生成并测试正则表达式，支持多种语言。", rating: 4.6, installed: false, icon: "terminal" },
	{ id: "t21", name: "3D 模型优化", verified: true, publisher: "由 Leros AI 提供", category: "视觉/媒体", tags: ["3D", "Optimize"], description: "智能减面优化 3D 模型，在保持视觉质量的同时大幅降低文件体积。", rating: 4.4, installed: false, icon: "palette" },
	{ id: "t22", name: "合同风险扫描", verified: false, publisher: "由 Leros AI 提供", category: "自然语言", tags: ["Risk", "Legal"], description: "扫描合同条款识别潜在法律风险点，标注风险等级并提供修改建议。", rating: 4.8, installed: false, icon: "description" },
	{ id: "t23", name: "异常流量检测", verified: true, publisher: "由 Leros AI 提供", category: "数据分析", tags: ["Security", "Traffic"], description: "实时监测流量模式，基于 ML 模型识别 DDoS、爬虫等异常访问行为。", rating: 4.5, installed: false, icon: "insights" },
	{ id: "t24", name: "Git 提交消息生成", verified: false, publisher: "由 Leros AI 提供", category: "代码生成", tags: ["Git", "Commit"], description: "分析代码 diff 自动生成规范的 commit message，支持 Conventional Commits。", rating: 4.7, installed: false, icon: "terminal" },
];

// ─── Component ───────────────────────────────────────────────────────────────

export function SkillMarketView() {
	const [activeCategory, setActiveCategory] = useState("");
	const [displayedSkills, setDisplayedSkills] = useState(() => [...MOCK_SKILLS]);
	const [hasMore, setHasMore] = useState(true);
	const scrollContainerRef = useRef<HTMLDivElement>(null);
	const templateIndexRef = useRef(0);

	const filteredSkills = displayedSkills.filter((s) =>
		activeCategory === "" ? true : s.category === activeCategory,
	);

	const loadMore = useCallback(() => {
		setDisplayedSkills((prev) => {
			const newBatch: SkillItem[] = [];
			const batchSize = 12;
			for (let i = 0; i < batchSize && templateIndexRef.current < MORE_SKILL_TEMPLATES.length; i++) {
				const tpl = MORE_SKILL_TEMPLATES[templateIndexRef.current]!;
				newBatch.push({
					id: `${tpl.id}-${Date.now()}-${i}`,
					name: tpl.name,
					verified: tpl.verified,
					publisher: tpl.publisher,
					category: tpl.category,
					tags: tpl.tags,
					description: tpl.description,
					rating: tpl.rating,
					installed: tpl.installed,
					icon: tpl.icon,
				});
				templateIndexRef.current++;
			}
			if (templateIndexRef.current >= MORE_SKILL_TEMPLATES.length) {
				setHasMore(false);
			}
			return [...prev, ...newBatch];
		});
	}, []);

	useEffect(() => {
		const container = scrollContainerRef.current;
		if (!container || !hasMore) return;

		const handleScroll = () => {
			const { scrollTop, scrollHeight, clientHeight } = container;
			// Fire when within 100px of the bottom
			if (scrollHeight - scrollTop - clientHeight < 100) {
				loadMore();
			}
		};

		container.addEventListener("scroll", handleScroll, { passive: true });
		return () => container.removeEventListener("scroll", handleScroll);
	}, [loadMore, hasMore]);

	return (
		<div
			data-slot="skill-market-view"
			className="flex min-h-0 h-full flex-1 flex-col bg-[var(--leros-app-bg)]"
		>
			{/* Header */}
			<div className="flex items-start justify-between border-b border-[var(--leros-control-border)] px-6 py-4">
				<div>
					<h1 className="text-2xl font-bold text-[var(--leros-text-strong)]">
						技能市场
					</h1>
					<p className="mt-1 text-sm text-[var(--leros-text-muted)]">
						探索并部署经过验证的技能，持续增强您的 AI 助手效能。
					</p>
				</div>
				<Button size="sm">
					<Plus className="size-4 mr-1" />
					创作技能
				</Button>
			</div>

			{/* Search + Filters */}
			<div className="flex items-center gap-4 border-b border-[var(--leros-control-border)] px-6 py-3">
				<div className="relative flex-1 max-w-xs">
					<Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-[var(--leros-text-subtle)]" />
					<input
						type="text"
						placeholder="搜索技能..."
						className="w-full rounded-md border border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] py-1.5 pl-7 pr-2 text-xs text-[var(--leros-text)] placeholder:text-[var(--leros-text-subtle)] focus:border-[var(--leros-primary)] focus:bg-white focus:outline-none transition-colors"
					/>
				</div>
				<div className="flex items-center gap-2 overflow-x-auto no-scrollbar">
					{CATEGORIES.map((cat) => {
						const isActive = activeCategory === cat.value;
						return (
							<button
								type="button"
								key={cat.value}
								onClick={() => setActiveCategory(cat.value)}
								className={cn(
									"whitespace-nowrap rounded-full border px-3.5 py-1 text-xs font-medium transition-colors shrink-0",
									isActive
										? "border-[var(--leros-primary)] bg-[var(--leros-primary-soft)] text-[var(--leros-primary)]"
										: "border-[var(--leros-control-border)] bg-transparent text-[var(--leros-text-muted)] hover:border-[var(--leros-text-subtle)] hover:text-[var(--leros-text)]",
								)}
							>
								{cat.label}
							</button>
						);
					})}
					<button
						type="button"
						className="flex items-center gap-1 whitespace-nowrap rounded-full border border-[var(--leros-control-border)] bg-transparent px-3.5 py-1 text-xs font-medium text-[var(--leros-text-muted)] hover:border-[var(--leros-text-subtle)] hover:text-[var(--leros-text)] transition-colors shrink-0"
					>
						<SlidersHorizontal className="size-3" />
						筛选
					</button>
				</div>
			</div>

			{/* Skill grid */}
			<div ref={scrollContainerRef} className="min-h-0 flex-1 overflow-y-auto px-6 py-8">
					<div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
						{filteredSkills.map((skill) => (
							<SkillCard key={skill.id} skill={skill} />
						))}

						{filteredSkills.length === 0 && (
							<div className="col-span-full flex flex-col items-center justify-center py-16 text-[var(--leros-text-subtle)]">
								<p className="text-sm">暂无符合条件的技能</p>
							</div>
						)}
					</div>

					{/* Infinite-scroll sentinel */}
					{hasMore && (
						<div className="flex justify-center py-8 text-xs text-[var(--leros-text-subtle)]">
							加载中...
						</div>
					)}
			</div>
		</div>
	);
}

// ─── Skill Card ──────────────────────────────────────────────────────────────

function SkillCard({ skill }: { skill: SkillItem }) {
	const [installed, setInstalled] = useState(skill.installed);

	return (
		<div
			className={cn(
				"group flex flex-col rounded-xl border border-[var(--leros-control-border)] bg-white p-4 transition-all duration-300",
				"hover:-translate-y-1 hover:border-[var(--leros-primary)] hover:shadow-lg",
			)}
		>
			{/* Top: icon + info + rating */}
			<div className="flex items-start justify-between mb-3">
				<div className="flex items-center gap-3">
					<div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[var(--leros-primary-soft)] text-[var(--leros-primary)] transition-all duration-300 group-hover:bg-[var(--leros-primary)] group-hover:text-white">
						{iconMap[skill.icon] ?? <Globe className="size-[18px]" />}
					</div>
					<div>
						<div className="flex items-center gap-1 mb-0.5">
							<h3 className="text-sm font-semibold text-[var(--leros-text-strong)] truncate max-w-[140px]">
								{skill.name}
							</h3>
							{skill.verified && (
								<span className="inline-flex shrink-0 text-[var(--leros-primary)]" title="已验证">
									<svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
										<path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" />
									</svg>
								</span>
							)}
						</div>
						<p className="text-[11px] text-[var(--leros-text-subtle)]">{skill.publisher}</p>
					</div>
				</div>
				<div className="flex shrink-0 items-center gap-1 rounded bg-amber-50 px-1.5 py-0.5 border border-amber-100">
					<Star className="size-3 fill-amber-500 text-amber-500" />
					<span className="text-[10px] font-bold text-amber-700">{skill.rating}</span>
				</div>
			</div>

			{/* Description */}
			<p className="flex-1 text-xs text-[var(--leros-text-muted)] mb-3 leading-relaxed line-clamp-2">
				{skill.description}
			</p>

			{/* Tags */}
			<div className="flex flex-wrap gap-1.5 mb-3">
				{skill.tags.map((tag) => (
					<span
						key={tag}
						className="px-2 py-0.5 rounded border border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] text-[10px] font-medium uppercase tracking-tight text-[var(--leros-text-muted)]"
					>
						{tag}
					</span>
				))}
			</div>

			{/* Bottom: install button */}
			<div className="flex items-center justify-end pt-3 border-t border-[var(--leros-control-border)] h-10">
				{installed ? (
					<span className="px-4 py-1 rounded-lg bg-[var(--leros-surface-soft)] text-xs font-medium text-[var(--leros-text-muted)] cursor-default">
						已安装
					</span>
				) : (
					<button
						type="button"
						className={cn(
							"rounded-lg bg-[var(--leros-primary)] px-4 py-1 text-xs font-medium text-white transition-all duration-200",
							"opacity-0 translate-y-1 group-hover:opacity-100 group-hover:translate-y-0",
							"hover:bg-[var(--leros-primary)]/90",
						)}
						onClick={() => setInstalled(true)}
					>
						安装技能
					</button>
				)}
			</div>
		</div>
	);
}
