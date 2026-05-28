import {
  infiniteQueryOptions,
  queryOptions,
  type QueryClient,
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
  type InfiniteData,
} from "@tanstack/react-query";
import { itemAPI, type Feed, type Item, type ListItemsParams } from "@/lib/api";
import {
  normalizeItemFilters,
  queryKeys,
  type ItemFilters,
  type NormalizedItemFilters,
} from "./keys";
import { usePreferencesStore } from "@/store";
import { useArticleSessionStore } from "@/store/article-session";

type ItemListResponse = Awaited<ReturnType<typeof itemAPI.list>>;
type ItemsInfiniteData = InfiniteData<ItemListResponse, number>;
type ItemsMutationContext = {
  prevItemLists: Array<[readonly unknown[], ItemsInfiniteData | undefined]>;
  prevItemDetails: Array<readonly [number, Item | undefined]>;
  prevFeeds: Feed[] | undefined;
  prevUnreadOverrides: Record<number, boolean | undefined>;
};
interface SetItemsReadStateOptions {
  keepReadItemsInUnreadLists?: boolean;
}

function buildListItemsParams(
  filters: NormalizedItemFilters,
  offset: number,
  pageSize: number,
): ListItemsParams {
  const params: ListItemsParams = {
    limit: pageSize,
    offset,
    order_by: "pub_date",
  };

  if (filters.feedId) params.feed_id = filters.feedId;
  if (filters.groupId) params.group_id = filters.groupId;
  if (filters.unread) params.unread = true;

  return params;
}

export const itemQueries = {
  list: (filters: ItemFilters, pageSize: number) => {
    const normalizedFilters = normalizeItemFilters(filters);

    return infiniteQueryOptions({
      queryKey: [...queryKeys.items.lists(), normalizedFilters, pageSize],
      queryFn: async ({ pageParam }) =>
        itemAPI.list(buildListItemsParams(normalizedFilters, pageParam, pageSize)),
      initialPageParam: 0,
      getNextPageParam: (lastPage, allPages) => {
        const fetched = allPages.reduce((n, p) => n + p.data.length, 0);
        return fetched < lastPage.total ? fetched : undefined;
      },
    });
  },
  detail: (itemId: number) =>
    queryOptions({
      queryKey: queryKeys.items.detail(itemId),
      queryFn: async () => {
        const res = await itemAPI.get(itemId);
        return res.data;
      },
    }),
};

export function useItems(filters: ItemFilters) {
  const articlePageSize = usePreferencesStore((state) => state.articlePageSize);
  return useInfiniteQuery(itemQueries.list(filters, articlePageSize));
}

export function useItem(itemId: number | null, enabled = true) {
  return useQuery({
    ...itemQueries.detail(itemId ?? 0),
    enabled: enabled && itemId !== null && itemId > 0,
  });
}

function snapshotItemsMutationState(
  qc: QueryClient,
  ids: number[],
): ItemsMutationContext {
  return {
    prevItemLists: qc.getQueriesData<ItemsInfiniteData>({
      queryKey: queryKeys.items.lists(),
    }),
    prevItemDetails: ids.map(
      (id) =>
        [id, qc.getQueryData<Item>(queryKeys.items.detail(id))] as const,
    ),
    prevFeeds: qc.getQueryData<Feed[]>(queryKeys.feeds.list()),
    prevUnreadOverrides: (() => {
      const overrides = useArticleSessionStore.getState().unreadOverrides;
      return Object.fromEntries(ids.map((id) => [id, overrides[id]]));
    })(),
  };
}

function applyOptimisticItemReadState(
  qc: QueryClient,
  ids: number[],
  targetUnread: boolean,
  prevFeeds: Feed[] | undefined,
  options: SetItemsReadStateOptions = {},
) {
  const idSet = new Set(ids);
  const feedDeltaMap = new Map<number, number>();
  const updatedItemsById = new Map<number, Item>();
  const changedIds = new Set<number>();

  const listEntries = qc.getQueriesData<ItemsInfiniteData>({
    queryKey: queryKeys.items.lists(),
  });

  for (const [queryKey] of listEntries) {
    const filters = queryKey[2] as NormalizedItemFilters | undefined;
    const removeReadItems =
      filters?.unread === true &&
      !targetUnread &&
      !options.keepReadItemsInUnreadLists;

    qc.setQueryData<ItemsInfiniteData>(queryKey, (old) => {
      if (!old) return old;

      let removedCount = 0;
      const pages = old.pages.map((page) => {
        const data: Item[] = [];

        for (const item of page.data) {
          if (!idSet.has(item.id) || item.unread === targetUnread) {
            data.push(item);
            continue;
          }

          if (!changedIds.has(item.id)) {
            const delta = targetUnread ? 1 : -1;
            feedDeltaMap.set(
              item.feed_id,
              (feedDeltaMap.get(item.feed_id) ?? 0) + delta,
            );
            changedIds.add(item.id);
          }

          const updatedItem = { ...item, unread: targetUnread };
          updatedItemsById.set(item.id, updatedItem);

          if (removeReadItems) {
            removedCount += 1;
            continue;
          }

          data.push(updatedItem);
        }

        return { ...page, data };
      });

      return {
        ...old,
        pages: pages.map((page) => ({
          ...page,
          total: removeReadItems
            ? Math.max(0, page.total - removedCount)
            : page.total,
        })),
      };
    });
  }

  for (const id of ids) {
    const optimisticItem = updatedItemsById.get(id);
    qc.setQueryData<Item>(queryKeys.items.detail(id), (old) =>
      old
        ? (() => {
            if (old.unread === targetUnread) return old;

            if (!changedIds.has(old.id)) {
              const delta = targetUnread ? 1 : -1;
              feedDeltaMap.set(
                old.feed_id,
                (feedDeltaMap.get(old.feed_id) ?? 0) + delta,
              );
              changedIds.add(old.id);
            }

            return { ...old, unread: targetUnread };
          })()
        : optimisticItem,
    );
  }

  if (prevFeeds && feedDeltaMap.size > 0) {
    qc.setQueryData(queryKeys.feeds.list(), (old: Feed[] | undefined) =>
      old?.map((feed) => {
        const delta = feedDeltaMap.get(feed.id) ?? 0;
        if (delta === 0) return feed;

        return {
          ...feed,
          unread_count: Math.max(0, feed.unread_count + delta),
        };
      }),
    );
  }
}

function rollbackItemsMutation(
  qc: QueryClient,
  context: ItemsMutationContext | undefined,
  targetUnread: boolean,
) {
  if (!context) return;

  for (const [key, data] of context.prevItemLists) {
    qc.setQueryData(key, data);
  }

  for (const [id, data] of context.prevItemDetails) {
    qc.setQueryData(queryKeys.items.detail(id), data);
  }

  if (context.prevFeeds) {
    qc.setQueryData(queryKeys.feeds.list(), context.prevFeeds);
  }

  useArticleSessionStore.setState((state) => {
    const unreadOverrides = { ...state.unreadOverrides };
    for (const [id, unread] of Object.entries(context.prevUnreadOverrides)) {
      const itemId = Number(id);
      if (unreadOverrides[itemId] !== targetUnread) {
        continue;
      }
      if (unread === undefined) {
        delete unreadOverrides[itemId];
      } else {
        unreadOverrides[itemId] = unread;
      }
    }
    return { unreadOverrides };
  });
}

function useSetItemsReadState(
  targetUnread: boolean,
  options: SetItemsReadStateOptions = {},
) {
  const qc = useQueryClient();
  const setUnreadOverride = useArticleSessionStore((s) => s.setUnreadOverride);
  const clearUnreadOverrides = useArticleSessionStore(
    (s) => s.clearUnreadOverrides,
  );

  return useMutation({
    mutationFn: async (ids: number[]) => {
      if (targetUnread) {
        await itemAPI.markUnread({ ids });
      } else {
        await itemAPI.markRead({ ids });
      }

      return ids;
    },
    onMutate: async (ids) => {
      const context = snapshotItemsMutationState(qc, ids);
      for (const id of ids) {
        setUnreadOverride(id, targetUnread);
      }
      applyOptimisticItemReadState(
        qc,
        ids,
        targetUnread,
        context.prevFeeds,
        options,
      );
      return context;
    },
    onError: (_error, _ids, context) => {
      rollbackItemsMutation(qc, context, targetUnread);
      void qc.invalidateQueries({ queryKey: queryKeys.items.all });
      void qc.invalidateQueries({ queryKey: queryKeys.feeds.all });
    },
    onSuccess: (ids) => {
      clearUnreadOverrides(ids);
    },
    onSettled: async () => {
      await qc.invalidateQueries({ queryKey: queryKeys.feeds.all });
    },
  });
}

export function useMarkItemsRead(options?: SetItemsReadStateOptions) {
  return useSetItemsReadState(false, options);
}

export function useMarkItemsUnread() {
  return useSetItemsReadState(true);
}
