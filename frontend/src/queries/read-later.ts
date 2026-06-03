import { useCallback, useMemo } from "react";
import {
  infiniteQueryOptions,
  type InfiniteData,
  useMutation,
  useInfiniteQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { readLaterAPI, type Item, type ReadLaterItem } from "@/lib/api";
import { useArticleSessionStore } from "@/store/article-session";
import { queryKeys } from "./keys";
import { useFeedLookup } from "./feeds";

const savedItemsPageSize = 100;

type ReadLaterListResponse = Awaited<ReturnType<typeof readLaterAPI.list>>;
type ReadLaterInfiniteData = InfiniteData<ReadLaterListResponse, number>;

function resolveReadLaterItemId(item: ReadLaterItem): number {
  return item.item_id ?? -item.id;
}

export const readLaterQueries = {
  list: () =>
    infiniteQueryOptions({
      queryKey: queryKeys.readLater.list(),
      queryFn: ({ pageParam }) =>
        readLaterAPI.list(savedItemsPageSize, pageParam),
      initialPageParam: 0,
      getNextPageParam: (lastPage, allPages) => {
        const fetched = allPages.reduce((n, page) => n + page.data.length, 0);
        return fetched < lastPage.total ? fetched : undefined;
      },
      staleTime: Number.POSITIVE_INFINITY,
    }),
};

export function useReadLaterItems() {
  const query = useInfiniteQuery(readLaterQueries.list());
  const readLaterItems = useMemo(
    () => query.data?.pages.flatMap((page) => page.data) ?? [],
    [query.data],
  );
  const total = query.data?.pages[0]?.total ?? readLaterItems.length;

  return { ...query, data: readLaterItems, total };
}

export function useReadLaterLookup() {
  const readLaterQuery = useReadLaterItems();
  const { data: readLaterItems = [], total } = readLaterQuery;
  const readLaterOverrides = useArticleSessionStore((s) => s.readLaterOverrides);

  const byArticleId = useMemo(
    () => new Map(readLaterItems.map((item) => [resolveReadLaterItemId(item), item])),
    [readLaterItems],
  );

  const isItemReadLater = useCallback(
    (itemId: number, fallback = false) =>
      readLaterOverrides[itemId] ?? (byArticleId.has(itemId) || fallback),
    [byArticleId, readLaterOverrides],
  );

  const getReadLaterByItemId = useCallback(
    (itemId: number) => byArticleId.get(itemId),
    [byArticleId],
  );

  return {
    readLaterItems,
    readLaterTotal: total,
    readLaterQuery,
    isItemReadLater,
    getReadLaterByItemId,
  };
}

export function useReadLaterArticles(filters: {
  feedId: number | null;
  groupId: number | null;
}) {
  const { readLaterItems } = useReadLaterLookup();
  const { feeds, getFeedById, getFeedsByGroup } = useFeedLookup();

  return useMemo(() => {
    let filtered = readLaterItems;

    if (filters.feedId) {
      const feed = getFeedById(filters.feedId);
      if (!feed) return [];
      filtered = filtered.filter((item) => item.feed_name === feed.name);
    } else if (filters.groupId) {
      const feedNames = new Set(getFeedsByGroup(filters.groupId).map((f) => f.name));
      filtered = filtered.filter((item) => feedNames.has(item.feed_name));
    }

    const feedIdByName = new Map(feeds.map((f) => [f.name, f.id]));

    return filtered.map(
      (item): Item => ({
        id: item.item_id ?? -item.id,
        feed_id: feedIdByName.get(item.feed_name) ?? 0,
        guid: item.link || `read-later:${item.id}`,
        title: item.title,
        link: item.link,
        content: item.content,
        pub_date: item.pub_date,
        unread: false,
        created_at: item.created_at,
        bookmarked: false,
        read_later: true,
      }),
    );
  }, [
    readLaterItems,
    filters.feedId,
    filters.groupId,
    feeds,
    getFeedById,
    getFeedsByGroup,
  ]);
}

export function useCreateReadLaterItem() {
  const qc = useQueryClient();
  const { getFeedById } = useFeedLookup();
  const setReadLaterOverride = useArticleSessionStore(
    (s) => s.setReadLaterOverride,
  );

  return useMutation({
    mutationFn: async (item: Item) => {
      const feed = getFeedById(item.feed_id);
      const res = await readLaterAPI.create({
        item_id: item.id,
        link: item.link,
        title: item.title,
        content: item.content,
        pub_date: item.pub_date,
        feed_name: feed?.name ?? "Unknown",
      });
      return res.data!;
    },
    onSuccess: (readLaterItem) => {
      const itemId = resolveReadLaterItemId(readLaterItem);
      setReadLaterOverride(itemId, true);
      void qc.invalidateQueries({ queryKey: queryKeys.readLater.all });
    },
  });
}

export function useDeleteReadLaterItem() {
  const qc = useQueryClient();
  const setReadLaterOverride = useArticleSessionStore(
    (s) => s.setReadLaterOverride,
  );

  return useMutation({
    mutationFn: async (readLaterItemId: number) => {
      await readLaterAPI.delete(readLaterItemId);
      return readLaterItemId;
    },
    onSuccess: (readLaterItemId) => {
      const readLaterItem = qc
        .getQueryData<ReadLaterInfiniteData>(queryKeys.readLater.list())
        ?.pages.flatMap((page) => page.data)
        .find((item) => item.id === readLaterItemId);
      if (readLaterItem) {
        setReadLaterOverride(resolveReadLaterItemId(readLaterItem), false);
      }

      void qc.invalidateQueries({ queryKey: queryKeys.readLater.all });
    },
  });
}

export function useDeleteReadLaterItemByItem() {
  const qc = useQueryClient();
  const setReadLaterOverride = useArticleSessionStore(
    (s) => s.setReadLaterOverride,
  );

  return useMutation({
    mutationFn: async (itemId: number) => {
      await readLaterAPI.deleteByItem(itemId);
      return itemId;
    },
    onSuccess: (itemId) => {
      setReadLaterOverride(itemId, false);
      void qc.invalidateQueries({ queryKey: queryKeys.readLater.all });
    },
  });
}
