import { useMemo, useCallback } from "react";
import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { feedAPI, type Feed, type RefreshJob } from "@/lib/api";
import { queryKeys } from "./keys";

const refreshJobPollIntervalMs = 1000;
const refreshJobPollTimeoutMs = 30 * 60 * 1000;

function sleep(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

async function waitForRefreshJob(job: RefreshJob): Promise<RefreshJob> {
  const deadline = Date.now() + refreshJobPollTimeoutMs;
  let current = job;

  while (current.status === "running") {
    if (Date.now() >= deadline) {
      throw new Error("refresh timed out");
    }

    await sleep(refreshJobPollIntervalMs);
    const res = await feedAPI.getRefreshJob(current.id);
    if (!res.data) {
      throw new Error("refresh job not found");
    }
    current = res.data;
  }

  if (current.status === "failed") {
    throw new Error(current.error || "refresh failed");
  }

  return current;
}

export const feedQueries = {
  list: () =>
    queryOptions({
      queryKey: queryKeys.feeds.list(),
      queryFn: async () => {
        const res = await feedAPI.list();
        return res.data;
      },
    }),
};

export function useFeeds() {
  return useQuery(feedQueries.list());
}

export function useFeedLookup() {
  const { data: feeds = [], isLoading } = useFeeds();

  const feedMap = useMemo(() => new Map(feeds.map((f) => [f.id, f])), [feeds]);

  const getFeedById = useCallback(
    (feedId: number) => feedMap.get(feedId),
    [feedMap],
  );

  const getFeedsByGroup = useCallback(
    (groupId: number) => feeds.filter((f) => f.group_id === groupId),
    [feeds],
  );

  return { feeds, getFeedById, getFeedsByGroup, isLoading };
}

export function useUnreadCounts() {
  const { data: feeds = [] } = useFeeds();

  const getUnreadCount = useCallback(
    (feedId: number) => feeds.find((f) => f.id === feedId)?.unread_count ?? 0,
    [feeds],
  );

  const getGroupUnreadCount = useCallback(
    (groupId: number) =>
      feeds
        .filter((f) => f.group_id === groupId)
        .reduce((sum, f) => sum + (f.unread_count ?? 0), 0),
    [feeds],
  );

  const getTotalUnreadCount = useCallback(
    () => feeds.reduce((sum, f) => sum + (f.unread_count ?? 0), 0),
    [feeds],
  );

  const getTotalItemCount = useCallback(
    () => feeds.reduce((sum, f) => sum + (f.item_count ?? 0), 0),
    [feeds],
  );

  return {
    getUnreadCount,
    getGroupUnreadCount,
    getTotalUnreadCount,
    getTotalItemCount,
  };
}

export function useCreateFeed() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: Parameters<typeof feedAPI.create>[0]) => {
      const res = await feedAPI.create(req);
      return res.data!;
    },
    onSuccess: (feed) => {
      qc.setQueryData(queryKeys.feeds.list(), (old: Feed[] | undefined) =>
        old ? [...old, feed] : [feed],
      );
    },
  });
}

export function useUpdateFeed() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      ...data
    }: Parameters<typeof feedAPI.update>[1] & { id: number }) => {
      const res = await feedAPI.update(id, data);
      return res.data!;
    },
    onSuccess: (updated) => {
      qc.setQueryData(queryKeys.feeds.list(), (old: Feed[] | undefined) =>
        old?.map((f) => (f.id === updated.id ? updated : f)),
      );
    },
  });
}

export function useDeleteFeed() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      await feedAPI.delete(id);
      return id;
    },
    onSuccess: (id) => {
      qc.setQueryData(queryKeys.feeds.list(), (old: Feed[] | undefined) =>
        old?.filter((f) => f.id !== id),
      );
      qc.invalidateQueries({ queryKey: queryKeys.items.all });
    },
  });
}

export function useRefreshFeeds() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const res = await feedAPI.refresh();
      if (!res.data) {
        throw new Error("refresh job not found");
      }
      return waitForRefreshJob(res.data);
    },
    onSuccess: async () => {
      await Promise.all([
        qc.invalidateQueries({ queryKey: queryKeys.feeds.all }),
        qc.invalidateQueries({ queryKey: queryKeys.items.all }),
      ]);
    },
  });
}

export function useRefreshFeed() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      const res = await feedAPI.refreshOne(id);
      if (!res.data) {
        throw new Error("refresh job not found");
      }
      return waitForRefreshJob(res.data);
    },
    onSuccess: async () => {
      await Promise.all([
        qc.invalidateQueries({ queryKey: queryKeys.feeds.all }),
        qc.invalidateQueries({ queryKey: queryKeys.items.all }),
      ]);
    },
  });
}

export function useMoveFeedsToGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      fromGroupId,
      toGroupId,
    }: {
      fromGroupId: number;
      toGroupId: number;
    }) => {
      const feeds = qc.getQueryData<Feed[]>(queryKeys.feeds.list()) ?? [];
      const toMove = feeds.filter((f) => f.group_id === fromGroupId);
      await Promise.all(
        toMove.map((f) => feedAPI.update(f.id, { group_id: toGroupId })),
      );
      return { fromGroupId, toGroupId };
    },
    onSuccess: ({ fromGroupId, toGroupId }) => {
      qc.setQueryData(queryKeys.feeds.list(), (old: Feed[] | undefined) =>
        old?.map((f) =>
          f.group_id === fromGroupId ? { ...f, group_id: toGroupId } : f,
        ),
      );
    },
  });
}
