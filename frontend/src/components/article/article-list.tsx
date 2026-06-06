import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { useVirtualizer } from "@tanstack/react-virtual";
import { CheckCheck, Loader2, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ArticleItem } from "./article-item";
import { ContentHeader } from "@/components/layout/content-header";
import { SidebarTrigger } from "@/components/layout/sidebar-trigger";
import { useArticleNavigation } from "@/hooks/use-keyboard";
import { useUrlState, type ArticleFilter } from "@/hooks/use-url-state";
import { useItems, useMarkItemsRead } from "@/queries/items";
import { useFeedLookup, useRefreshFeeds } from "@/queries/feeds";
import { useGroups } from "@/queries/groups";
import {
  useBookmarkLookup,
  useStarredItems,
} from "@/queries/bookmarks";
import {
  useReadLaterArticles,
  useReadLaterLookup,
} from "@/queries/read-later";
import { getFaviconUrl } from "@/lib/api/favicon";
import { dedupeArticlesByIdentity } from "@/lib/article-dedupe";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { usePreferencesStore } from "@/store";
import { useArticleSessionStore } from "@/store/article-session";
import type { Item } from "@/lib/api";

interface ArticleListProps {
  compact?: boolean;
}

export function ArticleList({ compact = false }: ArticleListProps) {
  const { t } = useI18n();
  const navigate = useNavigate();
  const {
    articleFilter,
    setArticleFilter,
    selectedFeedId,
    selectedGroupId,
    selectedArticleId,
    setSelectedArticle,
  } = useUrlState();
  const scrollViewportRef = useRef<HTMLDivElement>(null);
  const [starredUnreadOverrides, setStarredUnreadOverrides] = useState<
    Record<number, boolean>
  >({});
  const articlePageSize = usePreferencesStore((state) => state.articlePageSize);
  const unreadOverrides = useArticleSessionStore((state) => state.unreadOverrides);

  const isStarredMode = articleFilter === "starred";
  const isReadLaterMode = articleFilter === "read-later";
  const isSavedMode = isStarredMode || isReadLaterMode;

  // Items query for non-starred modes
  const itemsQuery = useItems({
    feedId: selectedFeedId,
    groupId: selectedGroupId,
    unread: articleFilter === "unread" ? true : undefined,
  });

  const { data: groups = [] } = useGroups();
  const { feeds, getFeedById, isLoading: isFeedsLoading } = useFeedLookup();
  const refreshFeeds = useRefreshFeeds();
  const markItemsRead = useMarkItemsRead();
  const { bookmarksQuery, getBookmarkByItemId } = useBookmarkLookup();
  const { getReadLaterByItemId, readLaterQuery } = useReadLaterLookup();

  // Flatten infinite query pages
  const items = useMemo(
    () => itemsQuery.data?.pages.flatMap((p) => p.data) ?? [],
    [itemsQuery.data],
  );

  const starredArticles = useStarredItems({
    feedId: selectedFeedId,
    groupId: selectedGroupId,
  });
  const readLaterArticles = useReadLaterArticles({
    feedId: selectedFeedId,
    groupId: selectedGroupId,
  });

  const sourceArticles = isStarredMode
    ? starredArticles
    : isReadLaterMode
      ? readLaterArticles
      : items;
  const filteredArticles = useMemo(() => {
    if (articleFilter !== "unread") return sourceArticles;

    return sourceArticles.filter(
      (article) =>
        (unreadOverrides[article.id] ?? article.unread) ||
        article.id === selectedArticleId,
    );
  }, [articleFilter, selectedArticleId, sourceArticles, unreadOverrides]);
  const articles = useMemo(() => {
    if (isSavedMode) return filteredArticles;

    return dedupeArticlesByIdentity(filteredArticles, selectedArticleId);
  }, [filteredArticles, isSavedMode, selectedArticleId]);
  const getArticleUnread = useCallback(
    (article: Item) => {
      const unreadOverride = unreadOverrides[article.id];
      if (unreadOverride !== undefined) return unreadOverride;

      if (!isSavedMode) return article.unread;

      const override = starredUnreadOverrides[article.id];
      if (override !== undefined) return override;

      return article.unread;
    },
    [isSavedMode, starredUnreadOverrides, unreadOverrides],
  );

  const displayArticles = useMemo(
    () =>
      articles.map((article) => {
        const unread = getArticleUnread(article);
        if (article.unread === unread) return article;

        return {
          ...article,
          unread,
        };
      }),
    [articles, getArticleUnread],
  );

  const hasMore = isStarredMode
    ? bookmarksQuery.hasNextPage
    : isReadLaterMode
      ? readLaterQuery.hasNextPage
      : itemsQuery.hasNextPage;
  const isLoading = isStarredMode
    ? bookmarksQuery.isLoading
    : isReadLaterMode
      ? readLaterQuery.isLoading
      : itemsQuery.isLoading;
  const isLoadingMore = isStarredMode
    ? bookmarksQuery.isFetchingNextPage
    : isReadLaterMode
      ? readLaterQuery.isFetchingNextPage
      : itemsQuery.isFetchingNextPage;
  const unreadDisplayCount = displayArticles.filter((a) => a.unread).length;
  const fetchNextPage = isStarredMode
    ? bookmarksQuery.fetchNextPage
    : isReadLaterMode
      ? readLaterQuery.fetchNextPage
      : itemsQuery.fetchNextPage;

  useEffect(() => {
    if (articleFilter !== "unread" || isSavedMode) return;
    if (
      !itemsQuery.hasNextPage ||
      itemsQuery.isFetching ||
      itemsQuery.isRefetching ||
      itemsQuery.isFetchingNextPage
    ) {
      return;
    }
    if (unreadDisplayCount >= articlePageSize) return;

    void fetchNextPage();
  }, [
    articleFilter,
    articlePageSize,
    fetchNextPage,
    isSavedMode,
    itemsQuery.data,
    itemsQuery.hasNextPage,
    itemsQuery.isFetching,
    itemsQuery.isFetchingNextPage,
    itemsQuery.isRefetching,
    unreadDisplayCount,
  ]);

  const articleIds = useMemo(
    () => displayArticles.map((article) => article.id),
    [displayArticles],
  );
  useArticleNavigation(articleIds, {
    enabled: selectedArticleId === null,
  });
  // TanStack Virtual intentionally returns mutable helpers for measurement.
  // eslint-disable-next-line react-hooks/incompatible-library
  const rowVirtualizer = useVirtualizer({
    count: displayArticles.length,
    getScrollElement: () => scrollViewportRef.current,
    estimateSize: () => (compact ? 96 : 116),
    overscan: 8,
    getItemKey: (index) => displayArticles[index]?.id ?? index,
  });

  // Determine title
  let title = t("article.list.all");
  if (selectedFeedId) {
    const feed = getFeedById(selectedFeedId);
    title = feed?.name ?? t("article.feedFallback");
  } else if (selectedGroupId) {
    const group = groups.find((g) => g.id === selectedGroupId);
    title = group?.name ?? t("article.groupFallback");
  }

  const unreadCount = unreadDisplayCount;
  const hasNoFeeds = !isFeedsLoading && feeds.length === 0;

  const handleMarkAllAsRead = async () => {
    const unreadIds = displayArticles
      .filter((a) => a.unread && a.id > 0)
      .map((a) => a.id);

    if (unreadIds.length === 0) return;

    try {
      await markItemsRead.mutateAsync(unreadIds);

      if (isSavedMode) {
        setStarredUnreadOverrides((prev) => {
          const next = { ...prev };
          for (const id of unreadIds) {
            next[id] = false;
          }
          return next;
        });
      }
    } catch (error) {
      console.error("Failed to mark all as read:", error);
    }
  };

  const handleRefreshFeeds = async () => {
    try {
      toast.info(t("feeds.toast.refreshing"));
      await refreshFeeds.mutateAsync();
      toast.success(t("feeds.toast.refreshed"));
    } catch (error) {
      console.error("Failed to refresh feeds:", error);
      toast.error(t("feeds.toast.refreshFailed"));
    }
  };

  return (
    <div className="flex h-full flex-col">
      <ContentHeader className={compact ? "px-3 sm:px-3" : undefined}>
        <div className="flex min-w-0 items-center gap-1">
          <SidebarTrigger />
          <h2 className="truncate text-[17px] font-semibold tracking-normal">
            {title}
          </h2>
          {!compact && unreadCount > 0 && (
            <span className="ml-1 shrink-0 rounded-full bg-primary/10 px-2 py-0.5 text-[11px] font-semibold text-primary">
              {unreadCount}
            </span>
          )}
        </div>
        <div className="mr-1 flex shrink-0 items-center gap-0.5">
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={handleMarkAllAsRead}
            disabled={unreadCount === 0}
            className="text-muted-foreground hover:text-foreground"
            aria-label={t("article.list.markAllRead")}
            title={t("article.list.markAllRead")}
          >
            <CheckCheck className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={handleRefreshFeeds}
            disabled={refreshFeeds.isPending}
            className="text-muted-foreground hover:text-foreground"
            aria-label={t("common.refresh")}
            title={t("common.refresh")}
          >
            <RefreshCw
              className={cn("h-4 w-4", refreshFeeds.isPending && "animate-spin")}
            />
          </Button>
        </div>
      </ContentHeader>

      {/* Article area with filter tabs */}
      <div
        className={cn(
          "flex min-h-0 flex-1 flex-col gap-2 overflow-hidden bg-list-panel px-3 py-2.5 sm:px-4",
          compact && "px-2 py-2 sm:px-2",
        )}
      >
        {!hasNoFeeds && (
          <Tabs
            value={articleFilter}
            onValueChange={(v) => setArticleFilter(v as ArticleFilter)}
            className="gap-0"
          >
            <TabsList className="h-7 w-full rounded-md border-border/70 bg-muted/55 p-0.5 shadow-none">
              <TabsTrigger value="all">{t("article.filter.all")}</TabsTrigger>
              <TabsTrigger value="unread">{t("article.filter.unread")}</TabsTrigger>
              <TabsTrigger value="starred">
                {t("article.filter.starred")}
              </TabsTrigger>
              <TabsTrigger value="read-later">
                {t("article.filter.readLater")}
              </TabsTrigger>
            </TabsList>
          </Tabs>
        )}

        {/* Article list */}
        <ScrollArea className="min-h-0 flex-1" viewportRef={scrollViewportRef}>
          <div className="space-y-0.5 pr-1">
            {isLoading && articles.length === 0 ? (
              <div className="space-y-1.5 p-1">
                {[1, 2, 3, 4, 5].map((i) => (
                  <div
                    key={i}
                    className="article-skeleton h-[72px] rounded-md"
                  />
                ))}
              </div>
            ) : articles.length === 0 ? (
              hasNoFeeds ? (
                <div className="flex flex-col items-center justify-center gap-3 py-12 text-center">
                  <p className="text-sm text-muted-foreground">
                    {t("article.list.noFeeds")}
                  </p>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => navigate({ to: "/feeds" })}
                  >
                    {t("article.list.openFeedManagement")}
                  </Button>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <p className="text-sm text-muted-foreground">
                    {t("article.list.noArticles")}
                  </p>
                </div>
              )
            ) : (
              <>
                <div
                  className="relative w-full"
                  style={{ height: `${rowVirtualizer.getTotalSize()}px` }}
                >
                  {rowVirtualizer.getVirtualItems().map((virtualRow) => {
                    const article = displayArticles[virtualRow.index];
                    if (!article) return null;

                    const feed = getFeedById(article.feed_id);
                    const bookmark = getBookmarkByItemId(article.id);
                    const readLaterItem = getReadLaterByItemId(article.id);

                    return (
                      <div
                        key={virtualRow.key}
                        ref={rowVirtualizer.measureElement}
                        data-index={virtualRow.index}
                        className="absolute top-0 left-0 w-full"
                        style={{
                          transform: `translateY(${virtualRow.start}px)`,
                        }}
                      >
                        <ArticleItem
                          article={article}
                          compact={compact}
                          isSelected={selectedArticleId === article.id}
                          onSelectArticle={setSelectedArticle}
                          feedName={
                            feed?.name ??
                            bookmark?.feed_name ??
                            readLaterItem?.feed_name ??
                            t("common.unknown")
                          }
                          feedFaviconUrl={
                            feed
                              ? getFaviconUrl(feed.link, feed.site_url)
                              : null
                          }
                        />
                      </div>
                    );
                  })}
                </div>
                {hasMore && (
                  <div className="flex justify-center py-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => fetchNextPage()}
                      disabled={isLoadingMore}
                      className="gap-2"
                    >
                      {isLoadingMore && (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      )}
                      {isLoadingMore
                        ? t("article.list.loading")
                        : t("article.list.loadMore")}
                    </Button>
                  </div>
                )}
              </>
            )}
          </div>
        </ScrollArea>
      </div>
    </div>
  );
}
