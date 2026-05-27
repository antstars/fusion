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
        className="w-full transform-gpu bg-background p-0 duration-200 sm:max-w-[min(760px,92vw)]"
        showCloseButton={false}
      >
        <SheetTitle className="sr-only">Article</SheetTitle>
        <ArticleDetailContent showCloseButton />
      </SheetContent>
    </Sheet>
  );
}
