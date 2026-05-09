import { FileText } from "lucide-react";
import { useUrlState } from "@/hooks/use-url-state";
import { useI18n } from "@/lib/i18n";
import { ArticleDetailContent } from "./article-detail-content";

export function ArticleDetailPane() {
  const { t } = useI18n();
  const { selectedArticleId } = useUrlState();

  if (selectedArticleId === null) {
    return (
      <section className="app-panel flex h-full flex-1 items-center justify-center rounded-lg border text-muted-foreground">
        <div className="flex flex-col items-center gap-3 text-center">
          <FileText className="h-10 w-10 opacity-50" />
          <p className="text-sm">{t("article.detail.selectArticle")}</p>
        </div>
      </section>
    );
  }

  return (
    <section className="app-panel h-full flex-1 overflow-hidden rounded-lg border">
      <ArticleDetailContent />
    </section>
  );
}
