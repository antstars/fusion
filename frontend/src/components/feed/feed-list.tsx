import { useLocation } from "@tanstack/react-router";
import { Clock, Inbox, Layers, Star } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { isArticleFilter } from "@/lib/article-filter";
import { useGroups } from "@/queries/groups";
import { useFeedLookup, useUnreadCounts } from "@/queries/feeds";
import { useBookmarkLookup } from "@/queries/bookmarks";
import { useReadLaterLookup } from "@/queries/read-later";
import { useUrlState } from "@/hooks/use-url-state";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { FeedGroup } from "./feed-group";
import { FeedItem } from "./feed-item";

export function FeedList() {
  const { t } = useI18n();
  const { data: groups = [], isLoading } = useGroups();
  const { feeds, getFeedsByGroup } = useFeedLookup();
  const { getTotalUnreadCount, getTotalItemCount } = useUnreadCounts();
  const { bookmarks } = useBookmarkLookup();
  const { readLaterItems } = useReadLaterLookup();
  const {
    selectedFeedId,
    selectedGroupId,
    articleFilter,
    selectTopLevelFilter,
  } = useUrlState();
  const { pathname } = useLocation();
  const firstPathSegment = pathname.split("/").filter(Boolean)[0];
  const isOnHomePage =
    typeof firstPathSegment === "string" && isArticleFilter(firstPathSegment);
  const isTopLevelSelected =
    isOnHomePage && selectedFeedId === null && selectedGroupId === null;
  const totalUnread = getTotalUnreadCount();
  const totalItems = getTotalItemCount();
  const starredCount = bookmarks.length;
  const readLaterCount = readLaterItems.length;

  const topFilters: Array<{
    value: "all" | "unread" | "starred" | "read-later";
    label: string;
    count: number;
    icon: typeof Inbox;
  }> = [
    {
      value: "unread",
      label: t("article.filter.unread"),
      count: totalUnread,
      icon: Inbox,
    },
    {
      value: "starred",
      label: t("article.filter.starred"),
      count: starredCount,
      icon: Star,
    },
    {
      value: "read-later",
      label: t("article.filter.readLater"),
      count: readLaterCount,
      icon: Clock,
    },
    {
      value: "all",
      label: t("article.filter.all"),
      count: totalItems,
      icon: Layers,
    },
  ];

  if (isLoading && groups.length === 0) {
    return (
      <div className="flex-1 p-4">
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-8 animate-pulse rounded-md bg-accent" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <ScrollArea className="min-h-0 flex-1 w-full min-w-0 overflow-hidden [&_[data-slot=scroll-area-viewport]>div]:!block">
      <div className="w-full min-w-0 space-y-1 px-2 pb-2">
        {/* Top-level filters */}
        <div className="space-y-0.5">
          {topFilters.map(({ value, label, count, icon: Icon }) => (
            <button
              key={value}
              onClick={() => selectTopLevelFilter(value)}
              className={cn(
                "sidebar-row flex h-8 w-full min-w-0 items-center gap-2 rounded-lg px-2 text-left text-[13px] transition-colors",
              )}
              data-selected={isTopLevelSelected && articleFilter === value}
            >
              <Icon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
              <span className="min-w-0 flex-1">{label}</span>
              <span className="sidebar-count shrink-0">{count}</span>
            </button>
          ))}
        </div>

        {/* Feeds header */}
        <div className="mt-3 flex items-center justify-between px-2 py-1">
          <span className="text-[11px] font-semibold uppercase tracking-[0.02em] text-muted-foreground">
            {t("search.group.feeds")}
          </span>
        </div>

        {/* Feed groups */}
        <div className="w-full min-w-0 space-y-0.5">
          {groups.map((group) => {
            const groupFeeds = getFeedsByGroup(group.id);

            return (
              <FeedGroup
                key={group.id}
                groupId={group.id}
                name={group.name}
                feeds={groupFeeds}
                showTotalCount={articleFilter === "all"}
              />
            );
          })}

          {/* Ungrouped feeds (group_id = 0) */}
          {feeds
            .filter((f) => f.group_id === 0)
            .map((feed) => (
              <FeedItem key={feed.id} feed={feed} />
            ))}
        </div>
      </div>
    </ScrollArea>
  );
}
