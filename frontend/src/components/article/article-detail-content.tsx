import { useEffect, useMemo, useRef } from "react";
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
import { useArticleNavigation } from "@/hooks/use-keyboard";
import { useUrlState } from "@/hooks/use-url-state";
import type { Item } from "@/lib/api";
import { getFaviconUrl } from "@/lib/api/favicon";
import { processArticleContent } from "@/lib/content";
import { useI18n } from "@/lib/i18n";
import { toSafeExternalUrl } from "@/lib/safe-url";
import { formatDate } from "@/lib/utils";
import {
  useBookmarkLookup,
  useCreateBookmark,
  useDeleteBookmark,
  useStarredItems,
} from "@/queries/bookmarks";
import { useFeedLookup } from "@/queries/feeds";
import {
  useItem,
  useItems,
  useMarkItemsRead,
  useMarkItemsUnread,
} from "@/queries/items";

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

export function useSelectedArticleDetail() {
  const {
    selectedArticleId,
    setSelectedArticle,
    setSelectedFeed,
    selectedFeedId,
    selectedGroupId,
    articleFilter,
  } = useUrlState();
  const { getFeedById } = useFeedLookup();
  const isStarredMode = articleFilter === "starred";

  const itemsQuery = useItems({
    feedId: selectedFeedId,
    groupId: selectedGroupId,
    unread: articleFilter === "unread" ? true : undefined,
  });
  const articles = useMemo(
    () => itemsQuery.data?.pages.flatMap((p) => p.data) ?? [],
    [itemsQuery.data],
  );
  const starredArticles = useStarredItems({
    feedId: selectedFeedId,
    groupId: selectedGroupId,
  });
  const sourceArticles = isStarredMode ? starredArticles : articles;
  const listArticles = useMemo(() => {
    if (articleFilter !== "unread") return sourceArticles;

    return sourceArticles.filter(
      (item) => item.unread || item.id === selectedArticleId,
    );
  }, [articleFilter, selectedArticleId, sourceArticles]);

  const markRead = useMarkItemsRead();
  const markUnread = useMarkItemsUnread();
  const autoReadItemIdsRef = useRef(new Set<number>());
  const { isItemStarred, getBookmarkByItemId } = useBookmarkLookup();
  const createBookmark = useCreateBookmark();
  const deleteBookmark = useDeleteBookmark();

  const articleIds = listArticles.map((article) => article.id);

  const storeArticle = selectedArticleId
    ? (listArticles.find((item) => item.id === selectedArticleId) ?? null)
    : null;

  const shouldFetchArticle =
    selectedArticleId !== null &&
    selectedArticleId > 0 &&
    (isStarredMode || storeArticle === null);
  const { data: fetchedArticle } = useItem(
    selectedArticleId,
    shouldFetchArticle,
  );

  const article: Item | null =
    (isStarredMode ? fetchedArticle ?? storeArticle : storeArticle ?? fetchedArticle) ??
    null;
  const canToggleRead =
    article !== null && article.id > 0 && (!isStarredMode || fetchedArticle !== undefined);
  const feed = article ? getFeedById(article.feed_id) : null;
  const bookmark = article ? getBookmarkByItemId(article.id) : null;
  const starred = article ? isItemStarred(article.id) : false;
  const safeArticleLink = article ? toSafeExternalUrl(article.link) : null;

  const handleToggleRead = async () => {
    if (!article || !canToggleRead) return;
    try {
      if (article.unread) {
        await markRead.mutateAsync([article.id]);
      } else {
        await markUnread.mutateAsync([article.id]);
      }
    } catch (error) {
      console.error("Failed to toggle read status:", error);
    }
  };

  const handleToggleStar = async () => {
    if (!article) return;
    try {
      if (starred) {
        const bookmark = getBookmarkByItemId(article.id);
        if (bookmark) {
          await deleteBookmark.mutateAsync(bookmark.id);
        }
      } else {
        await createBookmark.mutateAsync(article);
      }
    } catch (error) {
      console.error("Failed to toggle star:", error);
    }
  };

  const handleOpenOriginal = () => {
    if (!safeArticleLink) return;
    window.open(safeArticleLink, "_blank", "noopener,noreferrer");
  };

  const handleOpenFeed = () => {
    if (!article || article.feed_id <= 0) return;
    setSelectedFeed(article.feed_id);
  };

  useEffect(() => {
    if (!article || !canToggleRead || !article.unread) return;
    if (autoReadItemIdsRef.current.has(article.id)) return;

    autoReadItemIdsRef.current.add(article.id);
    void markRead.mutateAsync([article.id]).catch((error) => {
      autoReadItemIdsRef.current.delete(article.id);
      console.error("Failed to mark article as read:", error);
    });
  }, [article, canToggleRead, markRead]);

  const { goToNext, goToPrevious, hasNext, hasPrevious } =
    useArticleNavigation(articleIds, {
      enabled: selectedArticleId !== null,
      onToggleRead: () => {
        void handleToggleRead();
      },
      onToggleStar: () => {
        void handleToggleStar();
      },
      onOpenOriginal: handleOpenOriginal,
    });

  return {
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
  };
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
      <div className="flex items-center justify-between border-b px-4 py-3 sm:px-6">
        <div className="flex min-w-0 items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleToggleRead}
            disabled={!canToggleRead}
            className="h-auto gap-1.5 px-2.5 py-1.5 text-[13px] font-medium text-muted-foreground"
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
            className="h-auto gap-1.5 px-2.5 py-1.5 text-[13px] font-medium text-muted-foreground"
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
            className="h-auto gap-1.5 px-2.5 py-1.5 text-[13px] font-medium text-muted-foreground"
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

      <ScrollArea className="min-h-0 flex-1" viewportRef={scrollViewportRef}>
        <article className="mx-auto min-w-0 max-w-4xl px-5 py-6 sm:px-12 sm:py-8">
          <div className="space-y-3">
            <h1 className="text-[28px] leading-[1.3] font-bold">
              {article.title}
            </h1>
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
              {article.feed_id > 0 ? (
                <button
                  type="button"
                  onClick={handleOpenFeed}
                  className="flex max-w-48 items-center gap-1.5 rounded bg-muted px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent"
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
                <span className="flex max-w-48 items-center gap-1.5 rounded bg-muted px-2 py-1 text-xs font-medium text-muted-foreground">
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

      <div className="flex items-center justify-between border-t px-4 py-3 sm:px-6">
        <Button
          variant="outline"
          size="sm"
          onClick={goToPrevious}
          disabled={!hasPrevious()}
          className="h-auto gap-1.5 px-3 py-2 text-[13px] font-medium text-muted-foreground"
        >
          <ChevronLeft className="h-4 w-4" />
          {t("common.previous")}
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={goToNext}
          disabled={!hasNext()}
          className="h-auto gap-1.5 px-3 py-2 text-[13px] font-medium text-muted-foreground"
        >
          {t("common.next")}
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
