import { useEffect, useRef } from "react";
import {
  Circle,
  CircleCheck,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Star,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { FeedFavicon } from "@/components/feed/feed-favicon";
import { useSelectedArticleDetail } from "@/hooks/use-selected-article-detail";
import { getFaviconUrl } from "@/lib/api/favicon";
import { processArticleContent } from "@/lib/content";
import { useI18n } from "@/lib/i18n";
import { formatDate } from "@/lib/utils";

interface ArticleDetailContentProps {
  showCloseButton?: boolean;
}

function getLinkDomain(url: string) {
  try {
    return new URL(url).hostname;
  } catch {
    return url;
  }
}

export function ArticleDetailContent({
  showCloseButton = false,
}: ArticleDetailContentProps) {
  const { t } = useI18n();
  const scrollViewportRef = useRef<HTMLDivElement>(null);
  const {
    article,
    bookmark,
    canToggleRead,
    feed,
    goToNext,
    goToPrevious,
    handleOpenFeed,
    handleOpenOriginal,
    handleToggleRead,
    handleToggleStar,
    hasNext,
    hasPrevious,
    safeArticleLink,
    selectedArticleId,
    setSelectedArticle,
    starred,
  } = useSelectedArticleDetail();

  useEffect(() => {
    scrollViewportRef.current?.scrollTo({ top: 0 });
  }, [article?.id, selectedArticleId]);

  if (!article) {
    if (selectedArticleId !== null) {
      return (
        <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
          {t("article.list.loading")}
        </div>
      );
    }

    return null;
  }

  return (
    <div className="flex h-full flex-col">
      <div className="glass-panel-strong shrink-0 border-b">
        <div className="mx-auto flex h-14 w-full max-w-4xl items-center justify-between gap-3 px-5 sm:px-8">
          <div className="flex min-w-0 items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleToggleRead}
            disabled={!canToggleRead}
            className="glass-control h-auto gap-1.5 px-2.5 py-1.5 text-[13px] font-medium text-muted-foreground"
          >
            {article.unread ? (
              <Circle className="h-4 w-4 text-muted-foreground" />
            ) : (
              <CircleCheck className="h-4 w-4 text-primary" />
            )}
            {article.unread
              ? t("article.action.markRead")
              : t("article.action.markUnread")}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleToggleStar}
            className="glass-control h-auto gap-1.5 px-2.5 py-1.5 text-[13px] font-medium text-muted-foreground"
          >
            <Star
              className={`h-4 w-4 ${starred ? "fill-current text-amber-500" : ""}`}
            />
            {starred ? t("article.action.unstar") : t("article.action.star")}
          </Button>
          <Button
            asChild={Boolean(safeArticleLink)}
            variant="outline"
            size="sm"
            onClick={safeArticleLink ? undefined : handleOpenOriginal}
            disabled={!safeArticleLink}
            className="glass-control h-auto gap-1.5 px-2.5 py-1.5 text-[13px] font-medium text-muted-foreground"
          >
            {safeArticleLink ? (
              <a href={safeArticleLink} target="_blank" rel="noopener noreferrer">
                <ExternalLink className="h-4 w-4" />
                {t("article.action.original")}
              </a>
            ) : (
              <>
                <ExternalLink className="h-4 w-4" />
                {t("article.action.original")}
              </>
            )}
          </Button>
          </div>

          {showCloseButton && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => setSelectedArticle(null)}
              aria-label={t("common.cancel")}
            >
              <X className="h-[18px] w-[18px] text-muted-foreground" />
            </Button>
          )}
        </div>
      </div>

      <ScrollArea className="min-h-0 flex-1" viewportRef={scrollViewportRef}>
        <article className="mx-auto min-w-0 max-w-4xl px-5 py-8 sm:px-8 sm:py-10">
          <div className="space-y-3">
            <h1 className="text-[28px] leading-[1.3] font-bold">
              {article.title}
            </h1>
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
              {article.feed_id > 0 ? (
                <button
                  type="button"
                  onClick={handleOpenFeed}
                  className="glass-control flex max-w-48 items-center gap-1.5 rounded border px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent"
                >
                  {feed && (
                    <FeedFavicon
                      src={getFaviconUrl(feed.link, feed.site_url)}
                      className="h-3.5 w-3.5 rounded-sm"
                    />
                  )}
                  <span className="truncate hover:underline">
                    {feed?.name ?? bookmark?.feed_name ?? t("common.unknown")}
                  </span>
                </button>
              ) : (
                <span className="glass-control flex max-w-48 items-center gap-1.5 rounded border px-2 py-1 text-xs font-medium text-muted-foreground">
                  <span className="truncate">
                    {bookmark?.feed_name ?? t("common.unknown")}
                  </span>
                </span>
              )}
              <span className="text-muted-foreground">
                {formatDate(article.pub_date)}
              </span>
              {safeArticleLink ? (
                <a
                  href={safeArticleLink}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="truncate text-primary hover:underline"
                >
                  {getLinkDomain(safeArticleLink)}
                </a>
              ) : null}
            </div>
          </div>

          <div
            className="prose prose-neutral mt-6 min-w-0 max-w-none break-words dark:prose-invert"
            dangerouslySetInnerHTML={{
              __html: processArticleContent(
                article.content,
                safeArticleLink ?? undefined,
              ),
            }}
          />
        </article>
      </ScrollArea>

      <div className="glass-panel-strong shrink-0 border-t">
        <div className="mx-auto flex h-14 w-full max-w-4xl items-center justify-between px-5 sm:px-8">
          <Button
            variant="outline"
            size="sm"
            onClick={goToPrevious}
            disabled={!hasPrevious()}
            className="glass-control h-auto gap-1.5 px-3 py-2 text-[13px] font-medium text-muted-foreground"
          >
            <ChevronLeft className="h-4 w-4" />
            {t("common.previous")}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={goToNext}
            disabled={!hasNext()}
            className="glass-control h-auto gap-1.5 px-3 py-2 text-[13px] font-medium text-muted-foreground"
          >
            {t("common.next")}
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
