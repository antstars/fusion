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
        <div className="flex h-full min-w-0 gap-3 p-3">
          <div className="app-panel h-full w-[420px] flex-none overflow-hidden rounded-lg border">
            <ArticleList compact />
          </div>
          <ArticleDetailPane />
        </div>
      )}
    </AppLayout>
  );
}
