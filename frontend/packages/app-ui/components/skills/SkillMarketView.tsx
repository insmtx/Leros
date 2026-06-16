"use client";

import { Button } from "@leros/ui/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@leros/ui/components/ui/dropdown-menu";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@leros/ui/components/ui/tabs";
import { ChevronDown, Import, Plus } from "lucide-react";
import { useCallback, useState } from "react";
import type { SkillMarketplaceItem } from "@leros/store";
import { MarketplacePanel } from "./MarketplacePanel";
import { MySkillsPanel } from "./MySkillsPanel";

export function SkillMarketView() {
  const [activeTab, setActiveTab] = useState<"marketplace" | "mine">("marketplace");
  const [installedIds, setInstalledIds] = useState<Set<string>>(new Set());
  const [installingIds, setInstallingIds] = useState<Set<string>>(new Set());
  const [allItems, setAllItems] = useState<SkillMarketplaceItem[]>([]);

  const handleInstallSuccess = useCallback((skillId: string) => {
    setInstalledIds((prev) => new Set(prev).add(skillId));
  }, []);

  return (
    <div
      data-slot="skill-market-view"
      className="flex min-h-0 h-full flex-1 flex-col bg-[var(--leros-app-bg)]"
    >
      {/* Tabs — wraps the entire content area */}
      <Tabs
        value={activeTab}
        onValueChange={(v) => setActiveTab(v as "marketplace" | "mine")}
        className="min-h-0 flex-1 flex-col"
      >
        {/* Header with inline tabs */}
        <div className="flex items-start justify-between border-b border-[var(--leros-control-border)] px-6 py-4">
          <div>
            <TabsList variant="line" className="mb-3 -ml-1.5">
              <TabsTrigger
                value="marketplace"
                className="text-xl font-bold data-active:text-[var(--leros-text-strong)]"
              >
                技能市场
              </TabsTrigger>
              <TabsTrigger
                value="mine"
                className="text-xl font-bold data-active:text-[var(--leros-text-strong)]"
              >
                我的技能
              </TabsTrigger>
            </TabsList>
            <p className="text-sm text-[var(--leros-text-muted)]">
              {activeTab === "mine"
                ? "您已安装并正在使用的技能。"
                : "探索并部署经过验证的技能，持续增强您的 AI 助手效能。"}
            </p>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger
              render={(props) => (
                <Button size="sm" {...props}>
                  <Plus className="size-4 mr-1" />
                  创作技能
                  <ChevronDown className="size-3 ml-1 opacity-60" />
                </Button>
              )}
            />
            <DropdownMenuContent align="end" className="w-36">
              <DropdownMenuItem>
                <Plus className="size-4 mr-2" />
                创作技能
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Import className="size-4 mr-2" />
                导入技能
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Marketplace panel */}
        <TabsContent value="marketplace" className="min-h-0 flex-1 flex-col outline-none">
          <MarketplacePanel
            installedIds={installedIds}
            installingIds={installingIds}
            onInstallSuccess={handleInstallSuccess}
            onInstallingChange={setInstallingIds}
            onItemsLoaded={setAllItems}
          />
        </TabsContent>

        {/* My Skills panel */}
        <TabsContent value="mine" className="min-h-0 flex-1 overflow-y-auto px-6 py-8 outline-none">
          <MySkillsPanel installedIds={installedIds} allItems={allItems} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
