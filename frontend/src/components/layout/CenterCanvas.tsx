import {
  IconAt,
  IconCheck,
  IconChevronDown,
  IconChevronRight,
  IconCode,
  IconLoader,
  IconPaperclip,
  IconPlayerStop,
  IconSearch,
  IconSend,
  IconTerminal,
  IconX,
} from '@tabler/icons-react';
import { useCallback, useMemo, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { useLayoutStore } from '@/store/appStore';

type ToolCall = {
  id: string;
  name: string;
  input: string;
  output: string;
  status: 'running' | 'completed' | 'error';
};

type Message = {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: number;
  tools?: ToolCall[];
};

const mockMessages: Message[] = [
  {
    id: '1',
    role: 'user',
    content: '请帮我审查这个 PR：添加 OAuth 2.0 认证功能',
    timestamp: Date.now() - 120000,
  },
  {
    id: '2',
    role: 'assistant',
    content: '好的，我正在审查这个 Pull Request。让我先获取变更内容...',
    timestamp: Date.now() - 115000,
    tools: [
      {
        id: 't1',
        name: 'Bash',
        input: '{"command": "git diff main...feature/add-oauth"}',
        output: 'Found 12 changed files, +342 -89 lines',
        status: 'completed',
      },
      {
        id: 't2',
        name: 'Grep',
        input: '{"pattern": "TODO|FIXME", "include": "*.ts"}',
        output: 'Found 3 matches in 2 files',
        status: 'completed',
      },
      {
        id: 't3',
        name: 'Read',
        input: '{"filePath": "src/auth/oauth-provider.ts"}',
        output: 'Reading file (156 lines)...',
        status: 'completed',
      },
    ],
  },
  {
    id: '3',
    role: 'assistant',
    content: `## 代码审查完成

已完成对 12 个文件的审查，发现以下问题：

### 需要修改

1. **src/auth/oauth-provider.ts:45** - 缺少空值检查
   \`token\` 参数可能为 null，建议添加验证：
   \`\`\`ts
   if (!token) throw new Error('Token is required');
   \`\`\`

2. **src/hooks/useAuth.ts:78** - 内存泄漏风险
   useEffect 中未返回清理函数，可能导致组件卸载后仍在执行异步操作

### 建议改进

- **src/utils/jwt.ts** - 建议使用更安全的密钥管理方案
- **src/middleware/auth.ts** - 错误处理可以更详细

### ✅ 优点

- 代码结构清晰，模块化良好
- 测试覆盖率 87%
- 文档完善

建议修复以上问题后可以合并。`,
    timestamp: Date.now() - 30000,
  },
];

const modelOptions = ['GPT-4', 'Claude-3', 'DeepSeek'];

const formatTime = (timestamp: number) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  });
};

const getToolIcon = (toolName: string) => {
  switch (toolName) {
    case 'Bash':
      return IconTerminal;
    case 'Read':
    case 'Write':
      return IconCode;
    case 'Grep':
      return IconSearch;
    default:
      return IconCode;
  }
};

function ToolCallItem({ tool }: { tool: ToolCall }) {
  const [isOpen, setIsOpen] = useState(false);
  const Icon = getToolIcon(tool.name);

  const StatusIcon = {
    completed: IconCheck,
    error: IconX,
    running: IconLoader,
  }[tool.status];

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <div className="rounded-md border border-slate-200 bg-slate-50/50 overflow-hidden">
        <CollapsibleTrigger className="w-full">
          <div className="flex items-center gap-2 px-3 py-2 hover:bg-slate-100/50 transition-colors">
            {isOpen ? (
              <IconChevronDown className="size-3.5 text-slate-400" />
            ) : (
              <IconChevronRight className="size-3.5 text-slate-400" />
            )}
            <Icon className="size-3.5 text-slate-500" />
            <span className="text-xs font-medium text-slate-700">
              {tool.name}
            </span>
            <Badge
              variant={
                tool.status === 'completed'
                  ? 'default'
                  : tool.status === 'error'
                    ? 'destructive'
                    : 'secondary'
              }
              className="text-[10px] px-1.5 py-0"
            >
              <StatusIcon className="size-2.5" />
            </Badge>
          </div>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <div className="border-t border-slate-200 px-3 py-2 bg-white space-y-2">
            <div>
              <div className="text-[10px] uppercase tracking-wide text-slate-400 mb-1">
                Input
              </div>
              <pre className="text-xs font-mono text-slate-600 bg-slate-50 rounded px-2 py-1 overflow-x-auto">
                {tool.input}
              </pre>
            </div>
            {tool.output && (
              <div>
                <div className="text-[10px] uppercase tracking-wide text-slate-400 mb-1">
                  Output
                </div>
                <pre className="text-xs font-mono text-slate-600 bg-slate-50 rounded px-2 py-1 overflow-x-auto">
                  {tool.output}
                </pre>
              </div>
            )}
          </div>
        </CollapsibleContent>
      </div>
    </Collapsible>
  );
}

export function CenterCanvas() {
  const { activeConversationId, setInputFocused } = useLayoutStore(
    (state) => state,
  );
  const [inputValue, setInputValue] = useState('');
  const [isGenerating] = useState(false);
  const [selectedModel] = useState(modelOptions[0]);

  const handleSend = useCallback(() => {
    if (!inputValue.trim()) return;
    setInputValue('');
  }, [inputValue]);

  const handleInputFocus = useCallback(
    () => setInputFocused(true),
    [setInputFocused],
  );
  const handleInputBlur = useCallback(
    () => setInputFocused(false),
    [setInputFocused],
  );

  const messageElements = useMemo(
    () =>
      mockMessages.map((message) => (
        <div
          key={message.id}
          className={cn(
            'rounded-lg px-4 py-3',
            message.role === 'user'
              ? 'bg-white border border-slate-200 ml-8'
              : 'bg-transparent mr-8',
          )}
        >
          <div className="flex items-center gap-2 mb-2">
            <span className="text-xs font-medium text-slate-500 uppercase tracking-wide">
              {message.role === 'user' ? '你' : 'AI 助手'}
            </span>
            <span className="text-xs text-slate-400">
              {formatTime(message.timestamp)}
            </span>
          </div>
          <div
            className={cn(
              'text-sm leading-relaxed',
              message.role === 'assistant'
                ? 'font-serif text-slate-700'
                : 'text-slate-600',
            )}
          >
            {message.content.split('\n').map((line, index) => (
              <p
                key={`${message.id}-line-${index}`}
                className={line ? '' : 'mt-2'}
              >
                {line || '\u00A0'}
              </p>
            ))}
          </div>
          {message.tools && message.tools.length > 0 && (
            <div className="mt-3 space-y-2">
              {message.tools.map((tool) => (
                <ToolCallItem key={tool.id} tool={tool} />
              ))}
            </div>
          )}
        </div>
      )),
    [],
  );

  return (
    <div className="flex h-full flex-1 flex-col bg-slate-50">
      <div className="flex h-12 items-center justify-between border-b border-slate-200 bg-white px-6">
        <h1 className="text-sm font-medium text-slate-700">
          {activeConversationId ? '代码审查讨论' : '选择一个会话'}
        </h1>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" className="text-slate-500">
            分享
          </Button>
        </div>
      </div>

      <ScrollArea className="flex-1 min-h-0 overflow-hidden">
        <div className="mx-auto max-w-[650px] py-6 px-4">
          {!activeConversationId ? (
            <div className="flex flex-col items-center justify-center py-20 text-slate-400">
              <p>选择或创建一个会话开始对话</p>
            </div>
          ) : (
            <div className="space-y-6">{messageElements}</div>
          )}
        </div>
      </ScrollArea>

      <div className="border-t border-slate-200 bg-white">
        <div className="mx-auto max-w-[800px] p-4">
          <div className="relative rounded-lg border border-slate-200 bg-white shadow-sm">
            <textarea
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onFocus={handleInputFocus}
              onBlur={handleInputBlur}
              placeholder="输入消息... 使用 @ 提及，/ 命令"
              className="w-full resize-none rounded-lg px-4 py-3 text-sm min-h-[80px] max-h-[200px] focus:outline-none placeholder:text-slate-400"
              rows={1}
            />
            <div className="flex items-center justify-between border-t border-slate-100 px-3 py-2">
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  className="text-slate-400 hover:text-slate-600"
                >
                  <IconAt className="size-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  className="text-slate-400 hover:text-slate-600"
                >
                  <IconPaperclip className="size-4" />
                </Button>
                <button
                  type="button"
                  className="flex items-center gap-1 rounded-md px-2 py-1 text-xs text-slate-500 hover:bg-slate-100 transition-colors"
                >
                  {selectedModel}
                  <IconChevronDown className="size-3" />
                </button>
              </div>
              <div className="flex items-center gap-2">
                {isGenerating ? (
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-red-500 border-red-200 hover:bg-red-50"
                  >
                    <IconPlayerStop className="size-4 mr-1" />
                    停止
                  </Button>
                ) : (
                  <Button
                    size="sm"
                    className="bg-blue-500 hover:bg-blue-600 text-white"
                    onClick={handleSend}
                    disabled={!inputValue.trim()}
                  >
                    <IconSend className="size-4 mr-1" />
                    发送
                  </Button>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
