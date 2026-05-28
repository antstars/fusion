import { useEffect, useMemo, useRef } from "react";
import {
  useBookmarkLookup,
  useCreateBookmark,
  useDeleteBookmark,
  useStarredItems,
} from "@/queries/bookmarks";
import {
  useCreateReadLaterItem,
  useDeleteReadLaterItem,
  useReadLaterArticles,
  useReadLaterLookup,
} from "@/queries/read-later";
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
  const isReadLaterMode = articleFilter === "read-later";
  const isSavedMode = isStarredMode || isReadLaterMode;

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
  const readLaterArticles = useReadLaterArticles({
    feedId: selectedFeedId,
    groupId: selectedGroupId,
  });
  const sourceArticles = isStarredMode
    ? starredArticles
    : isReadLaterMode
      ? readLaterArticles
      : articles;
  const listArticles = useMemo(() => {
    if (articleFilter !== "unread") return sourceArticles;

    return sourceArticles.filter(
      (item) => item.unread || item.id === selectedArticleId,
    );
  }, [articleFilter, selectedArticleId, sourceArticles]);

  const markRead = useMarkItemsRead({ keepReadItemsInUnreadLists: true });
  const markUnread = useMarkItemsUnread();
  const autoReadItemIdsRef = useRef(new Set<number>());
  const { isItemStarred, getBookmarkByItemId } = useBookmarkLookup();
  const { isItemReadLater, getReadLaterByItemId } = useReadLaterLookup();
  const createBookmark = useCreateBookmark();
  const deleteBookmark = useDeleteBookmark();
  const createReadLaterItem = useCreateReadLaterItem();
  const deleteReadLaterItem = useDeleteReadLaterItem();

  const articleIds = listArticles.map((article) => article.id);

  const storeArticle = selectedArticleId
    ? (listArticles.find((item) => item.id === selectedArticleId) ?? null)
    : null;

  const shouldFetchArticle =
    selectedArticleId !== null &&
    selectedArticleId > 0 &&
    (isSavedMode || storeArticle === null);
  const { data: fetchedArticle } = useItem(
    selectedArticleId,
    shouldFetchArticle,
  );

  const article: Item | null =
    (isStarredMode
      ? (fetchedArticle ?? storeArticle)
      : isReadLaterMode
        ? (fetchedArticle ?? storeArticle)
      : (storeArticle ?? fetchedArticle)) ?? null;
  const canToggleRead =
    article !== null &&
    article.id > 0 &&
    (!isSavedMode || fetchedArticle !== undefined);
  const feed = article ? getFeedById(article.feed_id) : null;
  const bookmark = article ? getBookmarkByItemId(article.id) : null;
  const starred = article ? isItemStarred(article.id) : false;
  const readLaterItem = article ? getReadLaterByItemId(article.id) : null;
  const readLater = article ? isItemReadLater(article.id) : false;
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

  const handleToggleReadLater = async () => {
    if (!article) return;
    try {
      if (readLater) {
        const item = getReadLaterByItemId(article.id);
        if (item) {
          await deleteReadLaterItem.mutateAsync(item.id);
        }
      } else {
        await createReadLaterItem.mutateAsync(article);
      }
    } catch (error) {
      console.error("Failed to toggle read later:", error);
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
    handleToggleReadLater,
    handleToggleStar,
    hasNext,
    hasPrevious,
    readLater,
    readLaterItem,
    safeArticleLink,
    selectedArticleId,
    setSelectedArticle,
    starred,
  };
}
