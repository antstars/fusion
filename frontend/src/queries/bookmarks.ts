import { useCallback, useMemo } from "react";
import {
  infiniteQueryOptions,
  type InfiniteData,
  useMutation,
  useInfiniteQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { bookmarkAPI, type Bookmark, type Item } from "@/lib/api";
import { useArticleSessionStore } from "@/store/article-session";
import { queryKeys } from "./keys";
import { useFeedLookup } from "./feeds";

const savedItemsPageSize = 100;

type BookmarkListResponse = Awaited<ReturnType<typeof bookmarkAPI.list>>;
type BookmarkInfiniteData = InfiniteData<BookmarkListResponse, number>;

function resolveBookmarkItemId(bookmark: Bookmark): number {
  return bookmark.item_id ?? -bookmark.id;
}

export const bookmarkQueries = {
  list: () =>
    infiniteQueryOptions({
      queryKey: queryKeys.bookmarks.list(),
      queryFn: ({ pageParam }) =>
        bookmarkAPI.list(savedItemsPageSize, pageParam),
      initialPageParam: 0,
      getNextPageParam: (lastPage, allPages) => {
        const fetched = allPages.reduce((n, page) => n + page.data.length, 0);
        return fetched < lastPage.total ? fetched : undefined;
      },
      staleTime: Number.POSITIVE_INFINITY,
    }),
};

export function useBookmarks() {
  const query = useInfiniteQuery(bookmarkQueries.list());
  const bookmarks = useMemo(
    () => query.data?.pages.flatMap((page) => page.data) ?? [],
    [query.data],
  );
  const total = query.data?.pages[0]?.total ?? bookmarks.length;

  return { ...query, data: bookmarks, total };
}

export function useBookmarkLookup() {
  const bookmarksQuery = useBookmarks();
  const { data: bookmarks = [], total } = bookmarksQuery;
  const starredOverrides = useArticleSessionStore((s) => s.starredOverrides);

  const byArticleId = useMemo(
    () => new Map(bookmarks.map((b) => [resolveBookmarkItemId(b), b])),
    [bookmarks],
  );

  const isItemStarred = useCallback(
    (itemId: number, fallback = false) =>
      starredOverrides[itemId] ?? (byArticleId.has(itemId) || fallback),
    [byArticleId, starredOverrides],
  );

  const getBookmarkByItemId = useCallback(
    (itemId: number) => byArticleId.get(itemId),
    [byArticleId],
  );

  return {
    bookmarks,
    bookmarkTotal: total,
    bookmarksQuery,
    isItemStarred,
    getBookmarkByItemId,
  };
}

export function useStarredItems(filters: {
  feedId: number | null;
  groupId: number | null;
}) {
  const { bookmarks } = useBookmarkLookup();
  const { feeds, getFeedById, getFeedsByGroup } = useFeedLookup();

  return useMemo(() => {
    let filtered = bookmarks;

    if (filters.feedId) {
      const feed = getFeedById(filters.feedId);
      if (!feed) {
        return [];
      }
      filtered = filtered.filter((b) => b.feed_name === feed.name);
    } else if (filters.groupId) {
      const feedNames = new Set(getFeedsByGroup(filters.groupId).map((f) => f.name));
      filtered = filtered.filter((b) => feedNames.has(b.feed_name));
    }

    const feedIdByName = new Map(feeds.map((f) => [f.name, f.id]));

    return filtered.map(
      (bookmark): Item => ({
        id: bookmark.item_id ?? -bookmark.id,
        feed_id: feedIdByName.get(bookmark.feed_name) ?? 0,
        guid: bookmark.link || `bookmark:${bookmark.id}`,
        title: bookmark.title,
        link: bookmark.link,
        content: bookmark.content,
        pub_date: bookmark.pub_date,
        unread: bookmark.unread,
        created_at: bookmark.created_at,
        bookmarked: true,
        read_later: false,
      }),
    );
  }, [
    bookmarks,
    filters.feedId,
    filters.groupId,
    feeds,
    getFeedById,
    getFeedsByGroup,
  ]);
}

export function useCreateBookmark() {
  const qc = useQueryClient();
  const { getFeedById } = useFeedLookup();
  const setStarredOverride = useArticleSessionStore((s) => s.setStarredOverride);

  return useMutation({
    mutationFn: async (item: Item) => {
      const feed = getFeedById(item.feed_id);
      const res = await bookmarkAPI.create({
        item_id: item.id,
        link: item.link,
        title: item.title,
        content: item.content,
        pub_date: item.pub_date,
        feed_name: feed?.name ?? "Unknown",
      });
      return res.data!;
    },
    onSuccess: (bookmark) => {
      const itemId = resolveBookmarkItemId(bookmark);
      setStarredOverride(itemId, true);
      void qc.invalidateQueries({ queryKey: queryKeys.bookmarks.all });
    },
  });
}

export function useDeleteBookmark() {
  const qc = useQueryClient();
  const setStarredOverride = useArticleSessionStore((s) => s.setStarredOverride);

  return useMutation({
    mutationFn: async (bookmarkId: number) => {
      await bookmarkAPI.delete(bookmarkId);
      return bookmarkId;
    },
    onSuccess: (bookmarkId) => {
      const bookmark = qc
        .getQueryData<BookmarkInfiniteData>(queryKeys.bookmarks.list())
        ?.pages.flatMap((page) => page.data)
        .find((b) => b.id === bookmarkId);
      if (bookmark) {
        setStarredOverride(resolveBookmarkItemId(bookmark), false);
      }

      void qc.invalidateQueries({ queryKey: queryKeys.bookmarks.all });
    },
  });
}

export function useDeleteBookmarkByItem() {
  const qc = useQueryClient();
  const setStarredOverride = useArticleSessionStore((s) => s.setStarredOverride);

  return useMutation({
    mutationFn: async (itemId: number) => {
      await bookmarkAPI.deleteByItem(itemId);
      return itemId;
    },
    onSuccess: (itemId) => {
      setStarredOverride(itemId, false);
      void qc.invalidateQueries({ queryKey: queryKeys.bookmarks.all });
    },
  });
}
