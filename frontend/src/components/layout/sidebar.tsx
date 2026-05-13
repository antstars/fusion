import { Search, Settings, Rss } from "lucide-react";
import { useNavigate, useLocation } from "@tanstack/react-router";
import { FeedList } from "@/components/feed/feed-list";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { useUIStore } from "@/store";

export function Sidebar() {
  const { t } = useI18n();
  const { setSearchOpen, setSettingsOpen } = useUIStore();
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const isFeedsPage = pathname === "/feeds";

  return (
    <aside className="sidebar-typography app-panel flex h-full w-full flex-none flex-col overflow-hidden border-r text-sidebar-foreground shadow-none md:w-60">
      {/* Header */}
      <div className="flex h-14 shrink-0 items-center gap-2 border-b border-sidebar-border/70 px-4">
        <img
          src="/icon-32.png"
          alt={t("common.fusionLogo")}
          width={32}
          height={32}
          className="h-8 w-8 rounded-md"
        />
        <span className="text-base font-semibold">Fusion</span>
      </div>

      {/* Search button */}
      <div className="px-2 pt-3">
        <button
          className="flex h-10 w-full items-center justify-between rounded-md border border-sidebar-border bg-background px-3 text-muted-foreground transition-colors hover:bg-accent/70 hover:text-foreground"
          onClick={() => setSearchOpen(true)}
        >
          <div className="flex items-center gap-2">
            <Search className="h-4 w-4" />
            <span className="text-sm">{t("sidebar.search")}</span>
          </div>
          <kbd className="rounded bg-background/70 px-1.5 py-0.5 font-mono text-[11px] font-medium">
            Cmd+K / ?
          </kbd>
        </button>
      </div>

      {/* Feed list */}
      <FeedList />

      {/* Footer */}
      <div className="border-t border-sidebar-border/60 p-2">
        <button
          className={cn(
            "sidebar-row flex h-8 w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors",
          )}
          data-selected={isFeedsPage}
          onClick={() => navigate({ to: "/feeds" })}
        >
          <Rss className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span>{t("sidebar.manageFeeds")}</span>
        </button>
        <button
          className="sidebar-row flex h-8 w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors"
          onClick={() => setSettingsOpen(true)}
        >
          <Settings className="h-4 w-4 shrink-0 text-muted-foreground" />
          <span>{t("sidebar.settings")}</span>
        </button>
      </div>
    </aside>
  );
}
