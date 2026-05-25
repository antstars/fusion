import { ArticleDetailPane } from "@/components/article/article-detail-pane";
import { ArticleList } from "@/components/article/article-list";
import { AppLayout } from "@/components/layout/app-layout";
import { useIsMobile } from "@/hooks/use-mobile";

export function ArticlePage() {
  const isMobile = useIsMobile();

  return (
    <AppLayout>
      {isMobile ? (
        <ArticleList />
      ) : (
        <div className="flex h-full min-w-0 gap-0 bg-transparent p-0">
          <div className="h-full w-[420px] flex-none overflow-hidden border-r border-border bg-list-panel shadow-[inset_-1px_0_0_oklch(1_0_0_/_28%)]">
            <ArticleList compact />
          </div>
          <ArticleDetailPane />
        </div>
      )}
    </AppLayout>
  );
}
