import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ArrowUp,
  Circle,
  CircleCheck,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Clock,
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
  const titleBlockRef = useRef<HTMLDivElement>(null);
  const scrollFrameRef = useRef<number | null>(null);
  const [scrollProgress, setScrollProgress] = useState(0);
  const [showScrolledTitle, setShowScrolledTitle] = useState(false);
  const [canBackTop, setCanBackTop] = useState(false);
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
    handleToggleReadLater,
    handleToggleStar,
    hasNext,
    hasPrevious,
    readLater,
    safeArticleLink,
    selectedArticleId,
    setSelectedArticle,
    starred,
  } = useSelectedArticleDetail();

  const updateScrollState = useCallback(() => {
    const viewport = scrollViewportRef.current;
    if (!viewport) return;

    const maxScrollTop = Math.max(0, viewport.scrollHeight - viewport.clientHeight);
    const progress =
      maxScrollTop === 0
        ? 0
        : Math.round((viewport.scrollTop / maxScrollTop) * 100);
    const titleThreshold = titleBlockRef.current
      ? titleBlockRef.current.offsetTop + titleBlockRef.current.offsetHeight - 48
      : 96;

    setScrollProgress(Math.min(100, Math.max(0, progress)));
    setShowScrolledTitle(viewport.scrollTop > titleThreshold);
    setCanBackTop(viewport.scrollTop > 120);
  }, []);

  const scheduleScrollStateUpdate = useCallback(() => {
    if (scrollFrameRef.current !== null) return;

    scrollFrameRef.current = window.requestAnimationFrame(() => {
      scrollFrameRef.current = null;
      updateScrollState();
    });
  }, [updateScrollState]);

  useEffect(() => {
    scrollViewportRef.current?.scrollTo({ top: 0 });
    setScrollProgress(0);
    setShowScrolledTitle(false);
    setCanBackTop(false);
    window.requestAnimationFrame(updateScrollState);
  }, [article?.id, selectedArticleId, updateScrollState]);

  useEffect(() => {
    const viewport = scrollViewportRef.current;
    if (!viewport) return;

    updateScrollState();
    viewport.addEventListener("scroll", scheduleScrollStateUpdate, {
      passive: true,
    });

    return () => {
      viewport.removeEventListener("scroll", scheduleScrollStateUpdate);
      if (scrollFrameRef.current !== null) {
        window.cancelAnimationFrame(scrollFrameRef.current);
        scrollFrameRef.current = null;
      }
    };
  }, [article?.id, scheduleScrollStateUpdate, updateScrollState]);

  const handleBackTop = () => {
    scrollViewportRef.current?.scrollTo({ top: 0, behavior: "smooth" });
  };

  const articleHtml = useMemo(() => {
    if (!article) return "";

    return processArticleContent(article.content, safeArticleLink ?? undefined);
  }, [article?.content, safeArticleLink]);

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
    <div className="relative flex h-full flex-col">
      <div className="liquid-edge shrink-0 border-b border-border bg-panel">
        <div className="flex h-[52px] w-full items-center gap-3 px-4 sm:px-5">
          <div className="min-w-0 flex-1">
            <h2
              className={`truncate text-[15px] font-semibold transition-opacity duration-150 ${
                showScrolledTitle ? "opacity-100" : "opacity-0"
              }`}
              aria-hidden={!showScrolledTitle}
            >
              {article.title}
            </h2>
          </div>

          <div className="ml-auto flex shrink-0 items-center gap-0.5">
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={handleToggleRead}
              disabled={!canToggleRead}
              className="text-muted-foreground hover:text-foreground"
              aria-label={
                article.unread
                  ? t("article.action.markRead")
                  : t("article.action.markUnread")
              }
              title={
                article.unread
                  ? t("article.action.markRead")
                  : t("article.action.markUnread")
              }
            >
              {article.unread ? (
                <Circle className="h-4 w-4 text-muted-foreground" />
              ) : (
                <CircleCheck className="h-4 w-4 text-primary" />
              )}
            </Button>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={handleToggleStar}
              className="text-muted-foreground hover:text-foreground"
              aria-label={starred ? t("article.action.unstar") : t("article.action.star")}
              title={starred ? t("article.action.unstar") : t("article.action.star")}
            >
              <Star
                className={`h-4 w-4 ${starred ? "fill-current text-amber-500" : ""}`}
              />
            </Button>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={handleToggleReadLater}
              className="text-muted-foreground hover:text-foreground"
              aria-label={
                readLater
                  ? t("article.action.removeReadLater")
                  : t("article.action.readLater")
              }
              title={
                readLater
                  ? t("article.action.removeReadLater")
                  : t("article.action.readLater")
              }
            >
              <Clock
                className={`h-4 w-4 ${readLater ? "fill-current text-primary" : ""}`}
              />
            </Button>
            <Button
              asChild={Boolean(safeArticleLink)}
              variant="ghost"
              size="icon-sm"
              onClick={safeArticleLink ? undefined : handleOpenOriginal}
              disabled={!safeArticleLink}
              className="text-muted-foreground hover:text-foreground"
              aria-label={t("article.action.original")}
              title={t("article.action.original")}
            >
              {safeArticleLink ? (
                <a href={safeArticleLink} target="_blank" rel="noopener noreferrer">
                  <ExternalLink className="h-4 w-4" />
                </a>
              ) : (
                <ExternalLink className="h-4 w-4" />
              )}
            </Button>
          </div>

          {showCloseButton && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => setSelectedArticle(null)}
              aria-label={t("common.cancel")}
              className="shrink-0"
            >
              <X className="h-[18px] w-[18px] text-muted-foreground" />
            </Button>
          )}
        </div>
      </div>

      <ScrollArea
        className="mobile-smooth-scroll min-h-0 flex-1"
        viewportRef={scrollViewportRef}
      >
        <article className="mx-auto min-w-0 max-w-[760px] px-5 py-10 sm:px-10 sm:py-14">
          <div ref={titleBlockRef} className="space-y-4">
            <h1 className="text-[31px] leading-[1.18] font-semibold tracking-normal sm:text-[38px]">
              {article.title}
            </h1>
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm text-muted-foreground">
              {article.feed_id > 0 ? (
                <button
                  type="button"
                  onClick={handleOpenFeed}
                  className="flex max-w-48 items-center gap-1.5 rounded-md px-0 py-1 text-xs font-semibold text-primary transition-colors hover:underline"
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
                <span className="flex max-w-48 items-center gap-1.5 rounded-md py-1 text-xs font-semibold text-muted-foreground">
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
            className="prose prose-neutral mt-9 min-w-0 max-w-none break-words dark:prose-invert"
            dangerouslySetInnerHTML={{
              __html: articleHtml,
            }}
          />
        </article>
      </ScrollArea>

      <aside className="liquid-panel absolute top-24 right-7 z-10 hidden w-28 flex-col gap-2 rounded-xl border px-3 py-2 text-xs text-muted-foreground 2xl:flex">
        <div className="flex items-center gap-2">
          <div className="h-6 w-1 overflow-hidden rounded-full bg-muted">
            <div
              className="w-full rounded-full bg-primary transition-[height] duration-150"
              style={{ height: `${scrollProgress}%` }}
            />
          </div>
          <span className="font-medium text-foreground">{scrollProgress}%</span>
        </div>
        <button
          type="button"
          onClick={handleBackTop}
          disabled={!canBackTop}
          className="flex h-8 items-center gap-1.5 rounded-md text-left transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-40"
        >
          <ArrowUp className="h-3.5 w-3.5" />
          Back Top
        </button>
      </aside>

      <div className="liquid-edge shrink-0 border-t border-border bg-panel">
        <div className="mx-auto flex h-[48px] w-full max-w-[760px] items-center justify-between px-5 sm:px-10">
          <Button
            variant="ghost"
            size="sm"
            onClick={goToPrevious}
            disabled={!hasPrevious()}
            className="h-8 gap-1.5 px-2 text-[13px] font-medium text-muted-foreground"
          >
            <ChevronLeft className="h-4 w-4" />
            {t("common.previous")}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={goToNext}
            disabled={!hasNext()}
            className="h-8 gap-1.5 px-2 text-[13px] font-medium text-muted-foreground"
          >
            {t("common.next")}
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
