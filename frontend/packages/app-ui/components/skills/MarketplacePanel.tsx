"use client";

import { cn } from "@leros/ui/lib/utils";
import { Search, SlidersHorizontal } from "lucide-react";
import { toast } from "sonner";
import { useCallback, useEffect, useRef, useState } from "react";
import { skillMarketplaceApi, type SkillMarketplaceItem } from "@leros/store";
import { SkillCard } from "./SkillCard";

const CATEGORIES = [
  { value: "", label: "全部" },
  { value: "analysis", label: "数据分析" },
  { value: "language", label: "自然语言" },
  { value: "vision", label: "视觉/媒体" },
  { value: "code", label: "代码生成" },
];

const PAGE_SIZE = 80;

interface MarketplacePanelProps {
  installedIds: Set<string>;
  installingIds: Set<string>;
  onInstallSuccess: (skillId: string) => void;
  onInstallingChange: (value: Set<string> | ((prev: Set<string>) => Set<string>)) => void;
  onItemsLoaded: (items: SkillMarketplaceItem[]) => void;
}

export function MarketplacePanel({
  installedIds,
  installingIds,
  onInstallSuccess,
  onInstallingChange,
  onItemsLoaded,
}: MarketplacePanelProps) {
  const [items, setItems] = useState<SkillMarketplaceItem[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [keyword, setKeyword] = useState("");
  const [debouncedKeyword, setDebouncedKeyword] = useState("");
  const [activeCategory, setActiveCategory] = useState("");
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const loadingRef = useRef(false);
  const [mounted, setMounted] = useState(false);

  const handleInstall = useCallback(
    async (skill: SkillMarketplaceItem) => {
      const id = skill.skill_id;
      onInstallingChange(new Set(installingIds).add(id));
      try {
        await skillMarketplaceApi.install({
          source: skill.source_type,
          skill_id: skill.skill_id,
        });
        onInstallSuccess(id);
        toast.success("技能安装已提交");
      } catch (err: any) {
        const msg = err?.response?.data?.message ?? err?.message ?? "未知错误";
        toast.error(`安装失败：${msg}`);
      } finally {
        onInstallingChange((prev) => {
          const next = new Set(prev);
          next.delete(id);
          return next;
        });
      }
    },
    [installingIds, onInstallSuccess, onInstallingChange],
  );

  // debounce keyword
  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedKeyword(keyword), 300);
    return () => clearTimeout(timer);
  }, [keyword]);

  // fetch on keyword/category change (reset)
  useEffect(() => {
    let cancelled = false;
    const fetchItems = async () => {
      setLoading(true);
      try {
        const resp = await skillMarketplaceApi.search({
          keyword: debouncedKeyword || undefined,
          category: activeCategory || undefined,
          limit: PAGE_SIZE,
        });
        if (cancelled) return;
        const newItems = resp.data.data.items ?? [];
        setItems(newItems);
        onItemsLoaded(newItems);
        setHasMore(false);
      } catch (err) {
        if (!cancelled) console.error("Failed to fetch skills:", err);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    fetchItems();
    return () => {
      cancelled = true;
    };
  }, [debouncedKeyword, activeCategory, onItemsLoaded]);

  // load more (scroll trigger)
  const loadMore = useCallback(async () => {
    if (loadingRef.current || !hasMore) return;
    loadingRef.current = true;
    setLoadingMore(true);
    try {
      const resp = await skillMarketplaceApi.search({
        keyword: debouncedKeyword || undefined,
        category: activeCategory || undefined,
        limit: PAGE_SIZE,
      });
      const newItems = resp.data.data.items ?? [];
      if (newItems.length === 0) {
        setHasMore(false);
      } else {
        setItems((prev) => {
          const merged = [...prev, ...newItems];
          onItemsLoaded(merged);
          return merged;
        });
        setHasMore(false);
      }
    } catch (err) {
      console.error("Failed to load more skills:", err);
    } finally {
      setLoadingMore(false);
      loadingRef.current = false;
    }
  }, [debouncedKeyword, activeCategory, hasMore, onItemsLoaded]);

  // scroll listener
  useEffect(() => {
    const inner = scrollContainerRef.current;
    if (!inner) return;
    const container = inner.parentElement;
    if (!container) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      if (scrollHeight - scrollTop - clientHeight < 100) {
        loadMore();
      }
    };

    container.addEventListener("scroll", handleScroll, { passive: true });
    return () => container.removeEventListener("scroll", handleScroll);
  }, [loadMore]);

  return (
    <>
      {/* Search + Filters */}
      <div className="flex items-center gap-4 border-b border-[var(--leros-control-border)] px-6 py-3">
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-[var(--leros-text-subtle)]" />
          <input
            type="text"
            placeholder="搜索技能..."
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
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

      {/* Marketplace grid */}
      <div className="min-h-0 flex-1 overflow-y-auto px-6 py-8">
        <div ref={scrollContainerRef}>
          {!mounted || loading ? (
            <div className="flex items-center justify-center py-16 text-sm text-[var(--leros-text-subtle)]">
              加载中...
            </div>
          ) : items.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-[var(--leros-text-subtle)]">
              <p className="text-sm">暂无符合条件的技能</p>
            </div>
          ) : (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                {items.map((skill) => (
                  <SkillCard
                    key={skill.skill_id}
                    skill={skill}
                    onInstall={handleInstall}
                    installing={installingIds.has(skill.skill_id)}
                    installed={installedIds.has(skill.skill_id)}
                  />
                ))}
              </div>
              {loadingMore && (
                <div className="flex justify-center py-8 text-xs text-[var(--leros-text-subtle)]">
                  加载中...
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </>
  );
}
