import { FileText } from "lucide-react";
import { LiquidGlassPanel } from "@/components/ui/liquid-glass-panel";
import { useUrlState } from "@/hooks/use-url-state";
import { useI18n } from "@/lib/i18n";
import { ArticleDetailContent } from "./article-detail-content";

export function ArticleDetailPane() {
  const { t } = useI18n();
  const { selectedArticleId } = useUrlState();

  if (selectedArticleId === null) {
    return (
      <LiquidGlassPanel
        className="h-full flex-1 text-muted-foreground"
        contentClassName="items-center justify-center"
        cornerRadius={0}
      >
        <div className="flex flex-col items-center gap-3 text-center">
          <div className="flex size-14 items-center justify-center rounded-2xl bg-muted/55 shadow-[inset_0_1px_0_oklch(1_0_0_/_35%)]">
            <FileText className="h-7 w-7 opacity-50" />
          </div>
          <p className="text-sm">{t("article.detail.selectArticle")}</p>
        </div>
      </LiquidGlassPanel>
    );
  }

  return (
    <LiquidGlassPanel className="h-full flex-1" cornerRadius={0}>
      <ArticleDetailContent />
    </LiquidGlassPanel>
  );
}
