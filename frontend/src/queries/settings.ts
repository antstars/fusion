import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { settingsAPI, type UpdateRetentionSettingsRequest } from "@/lib/api";
import { queryKeys } from "./keys";

export function useRetentionSettings() {
  return useQuery({
    queryKey: queryKeys.settings.retention(),
    queryFn: async () => {
      const res = await settingsAPI.getRetention();
      return res.data;
    },
  });
}

export function useUpdateRetentionSettings() {
  const qc = useQueryClient();

  return useMutation({
    mutationFn: async (data: UpdateRetentionSettingsRequest) => {
      const res = await settingsAPI.updateRetention(data);
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.settings.all });
      qc.invalidateQueries({ queryKey: queryKeys.items.all });
      qc.invalidateQueries({ queryKey: queryKeys.feeds.all });
      qc.invalidateQueries({ queryKey: queryKeys.bookmarks.all });
      qc.invalidateQueries({ queryKey: queryKeys.readLater.all });
    },
  });
}
