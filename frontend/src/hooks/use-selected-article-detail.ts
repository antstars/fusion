import { useEffect, useMemo, useRef } from "react";
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
import type { Item } from "@/lib/api";
import { toSafeExternalUrl } from "@/lib/safe-url";
import { useArticleNavigation } from "./use-keyboard";
import { useUrlState } from "./use-url-state";
import { usePreferencesStore } from "@/store";

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
  const articlePageSize = usePreferencesStore((state) => state.articlePageSize);

  const itemsQuery = useItems({
    feedId: selectedFeedId,
    groupId: selectedGroupId,
    unread: articleFilter === "unread" ? true : undefined,
  });
  const fetchNextPage = itemsQuery.fetchNextPage;
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
  const unreadListCount = listArticles.filter((item) => item.unread).length;

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
    (isStarredMode
      ? (fetchedArticle ?? storeArticle)
      : (storeArticle ?? fetchedArticle)) ?? null;
  const canToggleRead =
    article !== null &&
    article.id > 0 &&
    (!isStarredMode || fetchedArticle !== undefined);
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
    if (articleFilter !== "unread" || isStarredMode) return;
    if (!itemsQuery.hasNextPage || itemsQuery.isFetchingNextPage) return;
    if (unreadListCount >= articlePageSize) return;
    if (itemsQuery.data && unreadListCount === 0) {
      const total = itemsQuery.data.pages.at(-1)?.total ?? 0;
      if (total === 0) return;
    }

    void fetchNextPage();
  }, [
    articleFilter,
    articlePageSize,
    fetchNextPage,
    isStarredMode,
    itemsQuery.data,
    itemsQuery.hasNextPage,
    itemsQuery.isFetchingNextPage,
    unreadListCount,
  ]);

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
