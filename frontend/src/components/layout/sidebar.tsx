import { Plus, Search, Settings, Rss } from "lucide-react";
import { useNavigate, useLocation } from "@tanstack/react-router";
import { FeedList } from "@/components/feed/feed-list";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { useUIStore } from "@/store";

export function Sidebar() {
  const { t } = useI18n();
  const { setAddFeedOpen, setSearchOpen, setSettingsOpen } = useUIStore();
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const isFeedsPage = pathname === "/feeds";

  return (
    <aside className="sidebar-typography flex h-full w-full flex-none flex-col overflow-hidden border-r border-sidebar-border bg-sidebar text-sidebar-foreground shadow-none md:w-[248px]">
      {/* Header */}
      <div className="flex h-[52px] shrink-0 items-center gap-2 bg-transparent px-3">
        <img
          src="/icon-32.png"
          alt={t("common.fusionLogo")}
          width={32}
          height={32}
          className="h-7 w-7 rounded-md"
        />
        <span className="min-w-0 flex-1 truncate text-[15px] font-semibold text-foreground">
          Fusion
        </span>
        <button
          type="button"
          className="liquid-control flex size-7 items-center justify-center rounded-md border text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground"
          onClick={() => setAddFeedOpen(true)}
          aria-label={t("feed.add.title")}
        >
          <Plus className="h-4 w-4" />
        </button>
      </div>

      {/* Search button */}
      <div className="bg-transparent px-2 pb-2">
        <button
          className="liquid-control flex h-8 w-full items-center justify-between rounded-lg border px-2.5 text-muted-foreground transition-colors hover:text-foreground"
          onClick={() => setSearchOpen(true)}
        >
          <div className="flex items-center gap-2">
            <Search className="h-3.5 w-3.5" />
            <span className="text-[13px]">{t("sidebar.search")}</span>
          </div>
          <kbd className="rounded-md bg-muted/70 px-1.5 py-0.5 font-mono text-[10px] font-medium text-muted-foreground shadow-[inset_0_1px_0_oklch(1_0_0_/_28%)]">
            ⌘K
          </kbd>
        </button>
      </div>

      {/* Feed list */}
      <FeedList />

      {/* Footer */}
      <div className="border-t border-sidebar-border bg-transparent p-2">
        <button
          className={cn(
            "sidebar-row flex h-8 w-full items-center gap-2 rounded-lg px-2 py-1.5 text-sm transition-colors",
          )}
          data-selected={isFeedsPage}
          onClick={() => navigate({ to: "/feeds" })}
        >
          <Rss className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span>{t("sidebar.manageFeeds")}</span>
        </button>
        <button
          className="sidebar-row flex h-8 w-full items-center gap-2 rounded-lg px-2 py-1.5 text-sm transition-colors"
          onClick={() => setSettingsOpen(true)}
        >
          <Settings className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span>{t("sidebar.settings")}</span>
        </button>
      </div>
    </aside>
  );
}
