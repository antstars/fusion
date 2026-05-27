import { ArticleDetailPane } from "@/components/article/article-detail-pane";
import { ArticleList } from "@/components/article/article-list";
import { AppLayout } from "@/components/layout/app-layout";
import { LiquidGlassPanel } from "@/components/ui/liquid-glass-panel";
import { useIsMobile } from "@/hooks/use-mobile";

export function ArticlePage() {
  const isMobile = useIsMobile();

  return (
    <AppLayout>
      {isMobile ? (
        <LiquidGlassPanel className="h-full w-full" cornerRadius={0}>
          <ArticleList />
        </LiquidGlassPanel>
      ) : (
        <div className="flex h-full min-w-0 gap-0 bg-transparent p-0">
          <LiquidGlassPanel
            className="h-full w-[420px] flex-none border-r border-border shadow-[inset_-1px_0_0_oklch(1_0_0_/_28%)]"
            cornerRadius={0}
          >
            <ArticleList compact />
          </LiquidGlassPanel>
          <ArticleDetailPane />
        </div>
      )}
    </AppLayout>
  );
}
