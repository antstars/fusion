import { lazy, Suspense, useEffect, useState, type ReactNode } from "react";
import { useLocation } from "@tanstack/react-router";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { Sidebar } from "./sidebar";
import { useKeyboardShortcuts } from "@/hooks/use-keyboard";
import { useI18n } from "@/lib/i18n";
import { useIsMobile } from "@/hooks/use-mobile";
import { useRefreshEvents } from "@/hooks/use-refresh-events";
import { useUrlState } from "@/hooks/use-url-state";
import { useUIStore } from "@/store/ui";

const ArticleDrawer = lazy(() =>
  import("@/components/article/article-drawer").then((module) => ({
    default: module.ArticleDrawer,
  })),
);
const SearchDialog = lazy(() =>
  import("@/components/search/search-dialog").then((module) => ({
    default: module.SearchDialog,
  })),
);
const SettingsDialog = lazy(() =>
  import("@/components/settings/settings-dialog").then((module) => ({
    default: module.SettingsDialog,
  })),
);
const AddGroupDialog = lazy(() =>
  import("@/components/group/add-group-dialog").then((module) => ({
    default: module.AddGroupDialog,
  })),
);
const AddFeedDialog = lazy(() =>
  import("@/components/feed/add-feed-dialog").then((module) => ({
    default: module.AddFeedDialog,
  })),
);
const EditFeedDialog = lazy(() =>
  import("@/components/feed/edit-feed-dialog").then((module) => ({
    default: module.EditFeedDialog,
  })),
);
const ImportOpmlDialog = lazy(() =>
  import("@/components/feed/import-opml-dialog").then((module) => ({
    default: module.ImportOpmlDialog,
  })),
);
const ShortcutsDialog = lazy(() =>
  import("@/components/layout/shortcuts-dialog").then((module) => ({
    default: module.ShortcutsDialog,
  })),
);

interface AppLayoutProps {
  children: ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const { t } = useI18n();
  const isMobile = useIsMobile();
  const isSidebarOpen = useUIStore((s) => s.isSidebarOpen);
  const setSidebarOpen = useUIStore((s) => s.setSidebarOpen);
  const isSearchOpen = useUIStore((s) => s.isSearchOpen);
  const isSettingsOpen = useUIStore((s) => s.isSettingsOpen);
  const isAddGroupOpen = useUIStore((s) => s.isAddGroupOpen);
  const isAddFeedOpen = useUIStore((s) => s.isAddFeedOpen);
  const isEditFeedOpen = useUIStore((s) => s.isEditFeedOpen);
  const isImportOpmlOpen = useUIStore((s) => s.isImportOpmlOpen);
  const isShortcutsOpen = useUIStore((s) => s.isShortcutsOpen);
  const { selectedArticleId } = useUrlState();
  const shouldRenderArticleDrawer = useHasBeenEnabled(
    isMobile && selectedArticleId !== null,
  );
  const shouldRenderSearchDialog = useHasBeenEnabled(isSearchOpen);
  const shouldRenderSettingsDialog = useHasBeenEnabled(isSettingsOpen);
  const shouldRenderAddGroupDialog = useHasBeenEnabled(isAddGroupOpen);
  const shouldRenderAddFeedDialog = useHasBeenEnabled(isAddFeedOpen);
  const shouldRenderEditFeedDialog = useHasBeenEnabled(isEditFeedOpen);
  const shouldRenderImportOpmlDialog = useHasBeenEnabled(isImportOpmlOpen);
  const shouldRenderShortcutsDialog = useHasBeenEnabled(isShortcutsOpen);

  useKeyboardShortcuts();
  useRefreshEvents();

  // Close mobile sidebar on navigation
  const location = useLocation();
  useEffect(() => {
    setSidebarOpen(false);
  }, [location.pathname, location.searchStr, setSidebarOpen]);

  return (
    <div className="flex h-screen w-full overflow-hidden bg-background/70">
      {/* Desktop sidebar */}
      {!isMobile && <Sidebar />}

      {/* Mobile sidebar */}
      {isMobile && (
        <Sheet open={isSidebarOpen} onOpenChange={setSidebarOpen}>
          <SheetContent
            side="left"
            className="w-[264px] transform-gpu border-sidebar-border bg-sidebar p-0 duration-200"
            showCloseButton={false}
          >
            <SheetTitle className="sr-only">{t("common.navigation")}</SheetTitle>
            <Sidebar />
          </SheetContent>
        </Sheet>
      )}

      {/* Main content */}
      <main className="min-w-0 flex-1 overflow-hidden bg-background/35">
        {children}
      </main>

      {/* Modals and drawers */}
      <Suspense fallback={null}>
        {isMobile && shouldRenderArticleDrawer && <ArticleDrawer />}
        {shouldRenderSearchDialog && <SearchDialog />}
        {shouldRenderSettingsDialog && <SettingsDialog />}
        {shouldRenderAddGroupDialog && <AddGroupDialog />}
        {shouldRenderAddFeedDialog && <AddFeedDialog />}
        {shouldRenderEditFeedDialog && <EditFeedDialog />}
        {shouldRenderImportOpmlDialog && <ImportOpmlDialog />}
        {shouldRenderShortcutsDialog && <ShortcutsDialog />}
      </Suspense>
    </div>
  );
}

function useHasBeenEnabled(enabled: boolean): boolean {
  const [hasBeenEnabled, setHasBeenEnabled] = useState(enabled);

  useEffect(() => {
    if (!enabled) return;

    const frame = window.requestAnimationFrame(() => setHasBeenEnabled(true));

    return () => window.cancelAnimationFrame(frame);
  }, [enabled]);

  return hasBeenEnabled;
}
