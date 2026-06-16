"use client";

import { cn } from "@leros/ui/lib/utils";
import { Loader2, Star } from "lucide-react";
import type { SkillMarketplaceItem } from "@leros/store";

interface SkillCardProps {
  skill: SkillMarketplaceItem;
  variant?: "marketplace" | "mine";
  onInstall?: (skill: SkillMarketplaceItem) => void;
  installing?: boolean;
  installed?: boolean;
  onUninstall?: (skill: SkillMarketplaceItem) => void;
  uninstalling?: boolean;
}

export function SkillCard({
  skill,
  variant = "marketplace",
  onInstall,
  installing,
  installed,
  onUninstall,
  uninstalling,
}: SkillCardProps) {
  const isLerosAI = skill.author === "Leros AI";
  const isMine = variant === "mine";

  return (
    <div
      className={cn(
        "group flex flex-col rounded-xl border border-[var(--leros-control-border)] bg-white p-4 transition-all duration-300",
        "hover:-translate-y-1 hover:border-[var(--leros-primary)] hover:shadow-lg",
      )}
    >
      {/* Top: avatar + info + rating */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3">
          {skill.icon ? (
            <img
              src={skill.icon}
              alt={skill.name}
              className="h-9 w-9 shrink-0 rounded-lg object-cover"
            />
          ) : (
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-[var(--leros-primary-soft)] text-[var(--leros-primary)] text-sm font-bold transition-all duration-300 group-hover:bg-[var(--leros-primary)] group-hover:text-white">
              {skill.name.charAt(0).toUpperCase()}
            </div>
          )}
          <div>
            <div className="flex items-center gap-1 mb-0.5">
              <h3 className="text-sm font-semibold text-[var(--leros-text-strong)] truncate max-w-[140px]">
                {skill.name}
              </h3>
              {isLerosAI && (
                <span
                  className="inline-flex shrink-0 text-[var(--leros-primary)]"
                  title="已验证"
                >
                  <svg
                    width="12"
                    height="12"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                  >
                    <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" />
                  </svg>
                </span>
              )}
            </div>
            <p className="text-[11px] text-[var(--leros-text-subtle)]">
              由 {skill.author || skill.source_type} 提供
            </p>
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-1 rounded bg-amber-50 px-1.5 py-0.5 border border-amber-100">
          <Star className="size-3 fill-amber-500 text-amber-500" />
          <span className="text-[10px] font-bold text-amber-700">4.5</span>
        </div>
      </div>

      {/* Description */}
      <p className="flex-1 text-xs text-[var(--leros-text-muted)] mb-3 leading-relaxed line-clamp-2">
        {skill.description}
      </p>

      {/* Tags + install count */}
      <div className="flex items-center gap-1.5 mb-3">
        <div className="flex flex-wrap gap-1.5 flex-1 min-w-0">
          {(skill.tags ?? []).map((tag: string) => (
            <span
              key={tag}
              className="px-2 py-0.5 rounded border border-[var(--leros-control-border)] bg-[var(--leros-surface-soft)] text-[10px] font-medium uppercase tracking-tight text-[var(--leros-text-muted)]"
            >
              {tag}
            </span>
          ))}
        </div>
        {!isMine && (
          <span className="shrink-0 text-[10px] text-[var(--leros-text-subtle)] ml-auto">
            {skill.installs} 安装
          </span>
        )}
      </div>

      {/* Bottom: install button / installed badge / uninstall button */}
      <div className="flex items-center justify-end pt-3 border-t border-[var(--leros-control-border)] h-10">
        {isMine ? (
          <>
            {/* Default: installed badge */}
            <span className="inline-flex items-center rounded-lg px-4 py-1 text-xs font-medium bg-green-50 text-green-600 border border-green-200 group-hover:hidden">
              已安装
            </span>
            {/* Hover: uninstall button */}
            <button
              type="button"
              disabled={uninstalling}
              onClick={() => onUninstall?.(skill)}
              className={cn(
                "hidden group-hover:inline-flex items-center gap-1.5 rounded-lg px-4 py-1 text-xs font-medium transition-all duration-200",
                "border border-red-200 text-red-600 hover:bg-red-50",
              )}
            >
              {uninstalling ? (
                <>
                  <Loader2 className="size-3 animate-spin" />
                  卸载中
                </>
              ) : (
                "卸载"
              )}
            </button>
          </>
        ) : (
          <button
            type="button"
            disabled={installing || installed}
            onClick={() => onInstall?.(skill)}
            className={cn(
              "inline-flex items-center gap-1.5 rounded-lg px-4 py-1 text-xs font-medium transition-all duration-200",
              "opacity-0 translate-y-1 group-hover:opacity-100 group-hover:translate-y-0",
              installed
                ? "bg-green-50 text-green-600 border border-green-200 cursor-default"
                : "bg-[var(--leros-primary)] text-white hover:bg-[var(--leros-primary)]/90",
              (installing || installed) && "opacity-100 translate-y-0",
            )}
          >
            {installing ? (
              <>
                <Loader2 className="size-3 animate-spin" />
                安装中
              </>
            ) : installed ? (
              "已安装"
            ) : (
              "安装技能"
            )}
          </button>
        )}
      </div>
    </div>
  );
}
