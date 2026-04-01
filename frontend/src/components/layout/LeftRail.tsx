import { IconPlus, IconTrash } from '@tabler/icons-react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { useLayoutStore } from '@/store/appStore';

export function LeftRail() {
  const {
    conversations,
    activeConversationId,
    switchConversation,
    createConversation,
    deleteConversation,
  } = useLayoutStore((state) => state);

  return (
    <div className="flex h-full w-[260px] flex-col border-r border-slate-200 bg-white">
      <div className="flex h-12 items-center justify-between border-b border-slate-200 px-4">
        <h2 className="text-sm font-medium tracking-wide uppercase text-slate-600">
          会话
        </h2>
        <Button
          variant="ghost"
          size="icon-sm"
          className="text-slate-500 hover:text-slate-700"
          onClick={() => createConversation('ws-1', '新会话')}
        >
          <IconPlus className="size-4" />
        </Button>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-2">
          {conversations.map((conversation) => (
            <button
              key={conversation.id}
              type="button"
              className={cn(
                'group relative flex items-center rounded-md px-2 py-1.5 text-sm cursor-pointer transition-colors w-full text-left',
                activeConversationId === conversation.id
                  ? 'bg-blue-50 text-blue-700'
                  : 'text-slate-700 hover:bg-slate-50',
              )}
              onClick={() => switchConversation(conversation.id)}
            >
              <span className="truncate flex-1">{conversation.title}</span>
              <Button
                variant="ghost"
                size="icon-xs"
                className="opacity-0 group-hover:opacity-100 transition-opacity text-slate-400 hover:text-red-500"
                onClick={(e) => {
                  e.stopPropagation();
                  deleteConversation(conversation.id);
                }}
              >
                <IconTrash className="size-3" />
              </Button>
            </button>
          ))}
          {conversations.length === 0 && (
            <p className="text-sm text-slate-400 text-center py-4">暂无会话</p>
          )}
        </div>
      </ScrollArea>

      <div className="border-t border-slate-200 p-2">
        <Button
          variant="ghost"
          size="sm"
          className="w-full justify-start text-slate-500"
        >
          <IconPlus className="size-4 mr-1.5" />
          新建工作区
        </Button>
      </div>
    </div>
  );
}
