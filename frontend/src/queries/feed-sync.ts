import { type InfiniteData, type QueryClient } from "@tanstack/react-query";
import type { Item, ListAPIResponse } from "@/lib/api";
import { queryKeys, type NormalizedItemFilters } from "./keys";

type ItemsInfiniteData = InfiniteData<ListAPIResponse<Item>, number>;

export async function refreshFeedSyncQueries(queryClient: QueryClient) {
  queryClient.removeQueries({
    queryKey: queryKeys.items.lists(),
    type: "inactive",
  });

  await Promise.all([
    queryClient.refetchQueries({
      queryKey: queryKeys.feeds.all,
      type: "active",
    }),
    queryClient.refetchQueries({
      queryKey: queryKeys.items.lists(),
      type: "active",
    }),
    queryClient.refetchQueries({
      queryKey: queryKeys.items.details(),
      type: "active",
    }),
  ]);

  pruneReadItemsFromUnreadLists(queryClient);
}

function pruneReadItemsFromUnreadLists(queryClient: QueryClient) {
  const listEntries = queryClient.getQueriesData<ItemsInfiniteData>({
    queryKey: queryKeys.items.lists(),
  });

  for (const [queryKey, old] of listEntries) {
    const filters = queryKey[2] as NormalizedItemFilters | undefined;
    if (filters?.unread !== true || !old) continue;

    let removedCount = 0;
    const pages = old.pages.map((page) => {
      const data = page.data.filter((item) => {
        if (item.unread) return true;

        removedCount += 1;
        return false;
      });

      return data.length === page.data.length ? page : { ...page, data };
    });

    if (removedCount === 0) continue;

    queryClient.setQueryData<ItemsInfiniteData>(queryKey, {
      ...old,
      pages: pages.map((page) => ({
        ...page,
        total: Math.max(0, page.total - removedCount),
      })),
    });
  }
}
