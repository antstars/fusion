import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { refreshFeedSyncQueries } from "@/queries/feed-sync";

const apiBase = (import.meta.env.VITE_API_BASE_URL || "/api").replace(/\/$/, "");
const refreshEventsURL = `${apiBase}/feeds/refresh-events`;
const refreshDelayMs = 200;

interface RefreshCompletedEvent {
  type: "refresh_completed";
  scope: "all" | "feed";
  feed_id: number | null;
  finished_at: number;
}

export function useRefreshEvents() {
  const queryClient = useQueryClient();
  const refreshTimerRef = useRef<number | null>(null);

  useEffect(() => {
    if (typeof window === "undefined" || !("EventSource" in window)) return;

    const eventSource = new EventSource(refreshEventsURL, {
      withCredentials: true,
    });

    const scheduleRefresh = () => {
      if (refreshTimerRef.current !== null) {
        window.clearTimeout(refreshTimerRef.current);
      }

      refreshTimerRef.current = window.setTimeout(() => {
        refreshTimerRef.current = null;
        void refreshFeedSyncQueries(queryClient);
      }, refreshDelayMs);
    };

    const handleRefreshCompleted = (event: Event) => {
      const message = event as MessageEvent<string>;
      try {
        const payload = JSON.parse(message.data) as RefreshCompletedEvent;
        if (payload.type !== "refresh_completed") return;
      } catch {
        return;
      }

      scheduleRefresh();
    };

    eventSource.addEventListener(refreshCompletedEventName, handleRefreshCompleted);

    return () => {
      eventSource.removeEventListener(
        refreshCompletedEventName,
        handleRefreshCompleted,
      );
      eventSource.close();

      if (refreshTimerRef.current !== null) {
        window.clearTimeout(refreshTimerRef.current);
        refreshTimerRef.current = null;
      }
    };
  }, [queryClient]);
}

const refreshCompletedEventName = "refresh-completed";
