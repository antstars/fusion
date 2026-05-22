import { useCallback, useMemo } from "react";
import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { readLaterAPI, type Item, type ReadLaterItem } from "@/lib/api";
import { useArticleSessionStore } from "@/store/article-session";
import { queryKeys } from "./keys";
import { useFeedLookup } from "./feeds";

function resolveReadLaterItemId(item: ReadLaterItem): number {
  return item.item_id ?? -item.id;
}

export const readLaterQueries = {
  list: () =>
    queryOptions({
      queryKey: queryKeys.readLater.list(),
      queryFn: async () => {
        const res = await readLaterAPI.list(100, 0);
        return res.data;
      },
      staleTime: Number.POSITIVE_INFINITY,
    }),
};

export function useReadLaterItems() {
  return useQuery(readLaterQueries.list());
}

export function useReadLaterLookup() {
  const { data: readLaterItems = [] } = useReadLaterItems();
  const readLaterOverrides = useArticleSessionStore((s) => s.readLaterOverrides);

  const byArticleId = useMemo(
    () => new Map(readLaterItems.map((item) => [resolveReadLaterItemId(item), item])),
    [readLaterItems],
  );

  const isItemReadLater = useCallback(
    (itemId: number) => readLaterOverrides[itemId] ?? byArticleId.has(itemId),
    [byArticleId, readLaterOverrides],
  );

  const getReadLaterByItemId = useCallback(
    (itemId: number) => byArticleId.get(itemId),
    [byArticleId],
  );

  return { readLaterItems, isItemReadLater, getReadLaterByItemId };
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
      qc.setQueryData(
        queryKeys.readLater.list(),
        (old: ReadLaterItem[] | undefined) => {
          if (!old) return [readLaterItem];

          const index = old.findIndex(
            (item) => resolveReadLaterItemId(item) === itemId,
          );
          if (index === -1) {
            return [readLaterItem, ...old];
          }

          const next = [...old];
          next[index] = readLaterItem;
          return next;
        },
      );
      setReadLaterOverride(itemId, true);
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
        .getQueryData<ReadLaterItem[]>(queryKeys.readLater.list())
        ?.find((item) => item.id === readLaterItemId);
      if (!readLaterItem) return;

      qc.setQueryData(
        queryKeys.readLater.list(),
        (old: ReadLaterItem[] | undefined) =>
          old?.filter((item) => item.id !== readLaterItemId),
      );
      setReadLaterOverride(resolveReadLaterItemId(readLaterItem), false);
    },
  });
}
