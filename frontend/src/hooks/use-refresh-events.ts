import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/queries/keys";

const apiBase = (import.meta.env.VITE_API_BASE_URL || "/api").replace(/\/$/, "");
const refreshEventsURL = `${apiBase}/feeds/refresh-events`;
const refreshInvalidationDelayMs = 200;

interface RefreshCompletedEvent {
  type: "refresh_completed";
  scope: "all" | "feed";
  feed_id: number | null;
  finished_at: number;
}

export function useRefreshEvents() {
  const queryClient = useQueryClient();
  const invalidationTimerRef = useRef<number | null>(null);

  useEffect(() => {
    if (typeof window === "undefined" || !("EventSource" in window)) return;

    const eventSource = new EventSource(refreshEventsURL, {
      withCredentials: true,
    });

    const scheduleInvalidation = () => {
      if (invalidationTimerRef.current !== null) {
        window.clearTimeout(invalidationTimerRef.current);
      }

      invalidationTimerRef.current = window.setTimeout(() => {
        invalidationTimerRef.current = null;
        void Promise.all([
          queryClient.invalidateQueries({ queryKey: queryKeys.feeds.all }),
          queryClient.invalidateQueries({ queryKey: queryKeys.items.all }),
        ]);
      }, refreshInvalidationDelayMs);
    };

    const handleRefreshCompleted = (event: Event) => {
      const message = event as MessageEvent<string>;
      try {
        const payload = JSON.parse(message.data) as RefreshCompletedEvent;
        if (payload.type !== "refresh_completed") return;
      } catch {
        return;
      }

      scheduleInvalidation();
    };

    eventSource.addEventListener(refreshCompletedEventName, handleRefreshCompleted);

    return () => {
      eventSource.removeEventListener(
        refreshCompletedEventName,
        handleRefreshCompleted,
      );
      eventSource.close();

      if (invalidationTimerRef.current !== null) {
        window.clearTimeout(invalidationTimerRef.current);
        invalidationTimerRef.current = null;
      }
    };
  }, [queryClient]);
}

const refreshCompletedEventName = "refresh-completed";
