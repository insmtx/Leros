export type ChatCommand = {
	code: string;
	label: string;
	description: string;
	keywords: string[];
};

export type MockAssistant = {
	code: string;
	name: string;
	description: string;
};

export const mockChatCommands: ChatCommand[] = [
	{
		code: "review",
		label: "代码审查",
		description: "检查代码问题、风险和改进建议",
		keywords: ["review", "code", "审查"],
	},
	{
		code: "summarize",
		label: "总结文档",
		description: "提炼内容重点并生成结构化摘要",
		keywords: ["summarize", "summary", "总结"],
	},
	{
		code: "explain",
		label: "解释代码",
		description: "说明代码逻辑、依赖和执行流程",
		keywords: ["explain", "code", "解释"],
	},
	{
		code: "test",
		label: "生成测试",
		description: "为给定功能补充测试方案或测试代码",
		keywords: ["test", "测试"],
	},
	{
		code: "assign",
		label: "需求指派",
		description: "分析需求并拆分可执行任务",
		keywords: ["assign", "task", "指派"],
	},
	{
		code: "ask",
		label: "知识问答",
		description: "基于已有上下文回答问题",
		keywords: ["ask", "knowledge", "问答"],
	},
];

export const mockAssistants: MockAssistant[] = [
	{
		code: "code-assistant",
		name: "代码助手",
		description: "擅长代码分析、实现和测试",
	},
	{
		code: "product-assistant",
		name: "产品助手",
		description: "擅长需求分析和产品规划",
	},
	{
		code: "knowledge-assistant",
		name: "知识助手",
		description: "擅长资料总结和知识问答",
	},
];
