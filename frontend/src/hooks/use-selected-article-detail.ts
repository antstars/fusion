import { useEffect, useMemo, useRef } from "react";
import {
  useBookmarkLookup,
  useCreateBookmark,
  useDeleteBookmark,
  useDeleteBookmarkByItem,
  useStarredItems,
} from "@/queries/bookmarks";
import {
  useCreateReadLaterItem,
  useDeleteReadLaterItem,
  useDeleteReadLaterItemByItem,
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
import { dedupeArticlesByIdentity } from "@/lib/article-dedupe";
import { toSafeExternalUrl } from "@/lib/safe-url";
import { useArticleNavigation } from "./use-keyboard";
import { useUrlState } from "./use-url-state";
import { useArticleSessionStore } from "@/store/article-session";

const autoMarkReadDelayMs = 800;

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
  const unreadOverrides = useArticleSessionStore((state) => state.unreadOverrides);

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
  const filteredArticles = useMemo(() => {
    if (articleFilter !== "unread") return sourceArticles;

    return sourceArticles.filter(
      (item) => unreadOverrides[item.id] ?? item.unread,
    );
  }, [articleFilter, sourceArticles, unreadOverrides]);
  const listArticles = useMemo(() => {
    if (isSavedMode) return filteredArticles;

    return dedupeArticlesByIdentity(filteredArticles, selectedArticleId);
  }, [filteredArticles, isSavedMode, selectedArticleId]);

  const markRead = useMarkItemsRead();
  const markUnread = useMarkItemsUnread();
  const pendingAutoReadItemIdsRef = useRef(new Set<number>());
  const manualUnreadHoldRef = useRef<number | null>(null);
  const { isItemStarred, getBookmarkByItemId } = useBookmarkLookup();
  const { isItemReadLater, getReadLaterByItemId } = useReadLaterLookup();
  const createBookmark = useCreateBookmark();
  const deleteBookmark = useDeleteBookmark();
  const deleteBookmarkByItem = useDeleteBookmarkByItem();
  const createReadLaterItem = useCreateReadLaterItem();
  const deleteReadLaterItem = useDeleteReadLaterItem();
  const deleteReadLaterItemByItem = useDeleteReadLaterItemByItem();

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
  const displayArticle: Item | null = article
    ? {
        ...article,
        unread: unreadOverrides[article.id] ?? article.unread,
      }
    : null;
  const canToggleRead =
    displayArticle !== null &&
    displayArticle.id > 0 &&
    (!isSavedMode || fetchedArticle !== undefined);
  const feed = displayArticle ? getFeedById(displayArticle.feed_id) : null;
  const bookmark = displayArticle ? getBookmarkByItemId(displayArticle.id) : null;
  const starred = displayArticle
    ? isItemStarred(displayArticle.id, displayArticle.bookmarked)
    : false;
  const readLaterItem = displayArticle
    ? getReadLaterByItemId(displayArticle.id)
    : null;
  const readLater = displayArticle
    ? isItemReadLater(displayArticle.id, displayArticle.read_later)
    : false;
  const safeArticleLink = displayArticle
    ? toSafeExternalUrl(displayArticle.link)
    : null;
  const displayArticleId = displayArticle?.id ?? null;

  const handleToggleRead = async () => {
    if (!displayArticle || !canToggleRead) return;
    try {
      if (displayArticle.unread) {
        await markRead.mutateAsync([displayArticle.id]);
      } else {
        manualUnreadHoldRef.current = displayArticle.id;
        await markUnread.mutateAsync([displayArticle.id]);
      }
    } catch (error) {
      console.error("Failed to toggle read status:", error);
    }
  };

  const handleToggleStar = async () => {
    if (!displayArticle) return;
    try {
      if (starred) {
        const bookmark = getBookmarkByItemId(displayArticle.id);
        if (bookmark) {
          await deleteBookmark.mutateAsync(bookmark.id);
        } else if (displayArticle.id > 0) {
          await deleteBookmarkByItem.mutateAsync(displayArticle.id);
        }
      } else {
        await createBookmark.mutateAsync(displayArticle);
      }
    } catch (error) {
      console.error("Failed to toggle star:", error);
    }
  };

  const handleToggleReadLater = async () => {
    if (!displayArticle) return;
    try {
      if (readLater) {
        const item = getReadLaterByItemId(displayArticle.id);
        if (item) {
          await deleteReadLaterItem.mutateAsync(item.id);
        } else if (displayArticle.id > 0) {
          await deleteReadLaterItemByItem.mutateAsync(displayArticle.id);
        }
      } else {
        await createReadLaterItem.mutateAsync(displayArticle);
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
    if (!displayArticle || displayArticle.feed_id <= 0) return;
    setSelectedFeed(displayArticle.feed_id);
  };

  useEffect(() => {
    if (
      manualUnreadHoldRef.current !== null &&
      manualUnreadHoldRef.current !== displayArticleId
    ) {
      manualUnreadHoldRef.current = null;
    }
  }, [displayArticleId]);

  useEffect(() => {
    if (selectedArticleId === null || selectedArticleId <= 0) {
      return;
    }
    if (manualUnreadHoldRef.current === selectedArticleId) return;
    if (pendingAutoReadItemIdsRef.current.has(selectedArticleId)) return;

    const articleId = selectedArticleId;
    const timer = window.setTimeout(() => {
      pendingAutoReadItemIdsRef.current.add(articleId);
      void markRead.mutateAsync([articleId]).catch((error) => {
        console.error("Failed to mark article as read:", error);
      }).finally(() => {
        pendingAutoReadItemIdsRef.current.delete(articleId);
      });
    }, autoMarkReadDelayMs);

    return () => window.clearTimeout(timer);
  }, [markRead, selectedArticleId]);

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
    article: displayArticle,
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
