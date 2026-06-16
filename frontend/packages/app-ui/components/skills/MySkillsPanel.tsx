"use client";

import { useEffect, useState } from "react";
import type { SkillMarketplaceItem } from "@leros/store";
import { SkillCard } from "./SkillCard";

interface MySkillsPanelProps {
  installedIds: Set<string>;
  allItems: SkillMarketplaceItem[];
}

export function MySkillsPanel({ installedIds, allItems }: MySkillsPanelProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const installedItems = allItems.filter((s) => installedIds.has(s.skill_id));

  if (!mounted) {
    return (
      <div className="flex items-center justify-center py-16 text-sm text-[var(--leros-text-subtle)]">
        加载中...
      </div>
    );
  }

  if (installedItems.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-[var(--leros-text-subtle)]">
        <p className="text-sm">暂无已安装的技能</p>
        <p className="text-xs mt-1">前往"技能市场"发现并安装技能</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      {installedItems.map((skill) => (
        <SkillCard key={skill.skill_id} skill={skill} variant="mine" />
      ))}
    </div>
  );
}
