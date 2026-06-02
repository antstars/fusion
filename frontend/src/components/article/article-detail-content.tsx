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

interface ReaderScrollState {
  scrollProgress: number;
  showScrolledTitle: boolean;
  canBackTop: boolean;
}

const initialReaderScrollState: ReaderScrollState = {
  scrollProgress: 0,
  showScrolledTitle: false,
  canBackTop: false,
};

function getLinkDomain(url: string) {
  try {
    return new URL(url).hostname;
  } catch {
    return url;
  }
}

function replaceFailedArticleImage(img: HTMLImageElement, label: string) {
  const fallback = document.createElement("div");
  fallback.className = "article-image-fallback";
  fallback.setAttribute("role", "img");
  fallback.setAttribute("aria-label", label);

  const text = document.createElement("span");
  text.textContent = label;
  fallback.appendChild(text);

  img.replaceWith(fallback);
}

export function ArticleDetailContent({
  showCloseButton = false,
}: ArticleDetailContentProps) {
  const { t } = useI18n();
  const scrollViewportRef = useRef<HTMLDivElement>(null);
  const articleContentRef = useRef<HTMLElement>(null);
  const titleBlockRef = useRef<HTMLDivElement>(null);
  const scrollFrameRef = useRef<number | null>(null);
  const scrollStateRef = useRef<ReaderScrollState>(initialReaderScrollState);
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

  const commitScrollState = useCallback((next: ReaderScrollState) => {
    const current = scrollStateRef.current;
    if (
      current.scrollProgress === next.scrollProgress &&
      current.showScrolledTitle === next.showScrolledTitle &&
      current.canBackTop === next.canBackTop
    ) {
      return;
    }

    scrollStateRef.current = next;
    if (current.scrollProgress !== next.scrollProgress) {
      setScrollProgress(next.scrollProgress);
    }
    if (current.showScrolledTitle !== next.showScrolledTitle) {
      setShowScrolledTitle(next.showScrolledTitle);
    }
    if (current.canBackTop !== next.canBackTop) {
      setCanBackTop(next.canBackTop);
    }
  }, []);

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

    commitScrollState({
      scrollProgress: Math.min(100, Math.max(0, progress)),
      showScrolledTitle: viewport.scrollTop > titleThreshold,
      canBackTop: viewport.scrollTop > 120,
    });
  }, [commitScrollState]);

  const scheduleScrollStateUpdate = useCallback(() => {
    if (scrollFrameRef.current !== null) return;

    scrollFrameRef.current = window.requestAnimationFrame(() => {
      scrollFrameRef.current = null;
      updateScrollState();
    });
  }, [updateScrollState]);

  useEffect(() => {
    scrollViewportRef.current?.scrollTo({ top: 0 });
    commitScrollState(initialReaderScrollState);
    scheduleScrollStateUpdate();
  }, [article?.id, commitScrollState, scheduleScrollStateUpdate, selectedArticleId]);

  useEffect(() => {
    const viewport = scrollViewportRef.current;
    if (!viewport) return;

    scheduleScrollStateUpdate();
    viewport.addEventListener("scroll", scheduleScrollStateUpdate, {
      passive: true,
    });
    window.addEventListener("resize", scheduleScrollStateUpdate);

    return () => {
      viewport.removeEventListener("scroll", scheduleScrollStateUpdate);
      window.removeEventListener("resize", scheduleScrollStateUpdate);
      if (scrollFrameRef.current !== null) {
        window.cancelAnimationFrame(scrollFrameRef.current);
        scrollFrameRef.current = null;
      }
    };
  }, [article?.id, scheduleScrollStateUpdate]);

  const articleContent = article?.content ?? "";

  const articleHtml = useMemo(() => {
    if (!articleContent) return "";

    return processArticleContent(articleContent, safeArticleLink ?? undefined);
  }, [articleContent, safeArticleLink]);

  useEffect(() => {
    const articleContent = articleContentRef.current;
    if (!articleContent) return;

    const handleImageLoad = (event: Event) => {
      const target = event.target;
      if (!(target instanceof HTMLImageElement)) return;
      if (!articleContent.contains(target)) return;

      scheduleScrollStateUpdate();
    };

    const handleImageError = (event: Event) => {
      const target = event.target;
      if (!(target instanceof HTMLImageElement)) return;
      if (!articleContent.contains(target)) return;

      replaceFailedArticleImage(target, t("article.imageUnavailable"));
      scheduleScrollStateUpdate();
    };

    articleContent.addEventListener("load", handleImageLoad, true);
    articleContent.addEventListener("error", handleImageError, true);
    for (const img of articleContent.querySelectorAll("img")) {
      if (img.complete) {
        scheduleScrollStateUpdate();
        break;
      }
    }

    return () => {
      articleContent.removeEventListener("load", handleImageLoad, true);
      articleContent.removeEventListener("error", handleImageError, true);
    };
  }, [articleHtml, scheduleScrollStateUpdate, t]);

  const handleBackTop = () => {
    scrollViewportRef.current?.scrollTo({ top: 0, behavior: "smooth" });
  };

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
    <div className="article-detail-shell relative flex h-full flex-col">
      <div className="shrink-0 border-b border-border/80 bg-panel">
        <div className="flex h-[48px] w-full items-center gap-2 px-3 sm:h-[52px] sm:px-5">
          {showCloseButton && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => setSelectedArticle(null)}
              aria-label={t("common.cancel")}
              className="shrink-0 sm:hidden"
            >
              <ChevronLeft className="h-[18px] w-[18px] text-muted-foreground" />
            </Button>
          )}
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

          <div className="ml-auto flex shrink-0 items-center gap-0.5 rounded-lg border border-border/55 bg-background/35 p-0.5">
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
                <Circle className="h-4 w-4 text-primary" />
              ) : (
                <CircleCheck className="h-4 w-4 text-muted-foreground" />
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
              className="hidden shrink-0 sm:inline-flex"
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
        <article
          ref={articleContentRef}
          className="mx-auto min-w-0 max-w-[740px] px-5 py-8 sm:px-10 sm:py-12"
        >
          <div ref={titleBlockRef} className="space-y-3">
            <h1 className="text-[28px] leading-[1.18] font-semibold tracking-normal sm:text-[36px]">
              {article.title}
            </h1>
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1.5 text-sm text-muted-foreground">
              {article.feed_id > 0 ? (
                <button
                  type="button"
                  onClick={handleOpenFeed}
                  className="flex max-w-48 items-center gap-1.5 rounded-md py-1 text-xs font-semibold text-primary transition-colors hover:underline"
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
            className="prose prose-neutral mt-8 min-w-0 max-w-none break-words dark:prose-invert"
            dangerouslySetInnerHTML={{
              __html: articleHtml,
            }}
          />
        </article>
      </ScrollArea>

      <aside
        className="reader-progress-control app-panel"
        aria-label="Reading progress"
      >
        <div
          className="h-20 w-1.5 overflow-hidden rounded-full bg-muted/80"
          aria-hidden="true"
        >
          <div
            className="w-full rounded-full bg-primary transition-[height] duration-200 ease-out"
            style={{ height: `${scrollProgress}%` }}
          />
        </div>
        <button
          type="button"
          onClick={handleBackTop}
          disabled={!canBackTop}
          className="flex size-8 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:pointer-events-none disabled:opacity-35"
          aria-label="Back to top"
          title="Back to top"
        >
          <ArrowUp className="h-4 w-4" />
        </button>
      </aside>

      <div className="shrink-0 border-t border-border bg-panel">
        <div className="mx-auto flex h-[44px] w-full max-w-[740px] items-center justify-between px-4 sm:h-[48px] sm:px-10">
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
