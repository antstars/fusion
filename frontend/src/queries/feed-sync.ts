import type { QueryClient } from "@tanstack/react-query";
import { queryKeys } from "./keys";

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
  ]);
}
