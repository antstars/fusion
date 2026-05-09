import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { useUrlState } from "@/hooks/use-url-state";
import { ArticleDetailContent } from "./article-detail-content";

export function ArticleDrawer() {
  const { selectedArticleId, setSelectedArticle } = useUrlState();

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setSelectedArticle(null);
    }
  };

  return (
    <Sheet open={selectedArticleId !== null} onOpenChange={handleOpenChange}>
      <SheetContent
        side="right"
        className="glass-panel-strong w-full p-0 sm:max-w-[max(720px,50vw)]"
        showCloseButton={false}
      >
        <SheetTitle className="sr-only">Article</SheetTitle>
        <ArticleDetailContent showCloseButton />
      </SheetContent>
    </Sheet>
  );
}
