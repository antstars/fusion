import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { LiquidGlassPanel } from "@/components/ui/liquid-glass-panel";
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
        className="w-full transform-gpu bg-background p-0 duration-200 sm:max-w-[max(720px,50vw)]"
        showCloseButton={false}
      >
        <SheetTitle className="sr-only">Article</SheetTitle>
        <LiquidGlassPanel className="h-full w-full" cornerRadius={0}>
          <ArticleDetailContent showCloseButton />
        </LiquidGlassPanel>
      </SheetContent>
    </Sheet>
  );
}
