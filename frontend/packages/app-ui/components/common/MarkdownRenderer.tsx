import Markdown from "react-markdown";
import rehypeKatex from "rehype-katex";
import remarkGfm from "remark-gfm";
import remarkMath from "remark-math";

type MarkdownRendererProps = {
	content: string;
	className?: string;
};

export function MarkdownRenderer({ content, className }: MarkdownRendererProps) {
	return (
		<div className={className}>
			{/* 统一挂载 markdown、GFM 和数学公式插件，避免不同入口的渲染能力不一致。 */}
			<Markdown remarkPlugins={[remarkGfm, remarkMath]} rehypePlugins={[rehypeKatex]}>
				{content}
			</Markdown>
		</div>
	);
}
