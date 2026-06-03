import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { useUrlState } from "@/hooks/use-url-state";
import { useI18n } from "@/lib/i18n";
import { ArticleDetailContent } from "./article-detail-content";

export function ArticleDrawer() {
  const { t } = useI18n();
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
        <SheetTitle className="sr-only">{t("article.detail.title")}</SheetTitle>
        <ArticleDetailContent showCloseButton />
      </SheetContent>
    </Sheet>
  );
}
