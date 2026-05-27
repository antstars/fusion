import { useState } from "react";
import { ChevronRight } from "lucide-react";
import { Collapsible, CollapsibleContent } from "@/components/ui/collapsible";
import { cn } from "@/lib/utils";
import { useUrlState } from "@/hooks/use-url-state";
import { FeedItem } from "./feed-item";
import type { Feed } from "@/lib/api";

interface FeedGroupProps {
  groupId: number;
  name: string;
  feeds: Feed[];
  showTotalCount?: boolean;
}

export function FeedGroup({
  groupId,
  name,
  feeds,
  showTotalCount = false,
}: FeedGroupProps) {
  const [isOpen, setIsOpen] = useState(false);
  const { selectedGroupId, setSelectedGroup } = useUrlState();
  const isSelected = selectedGroupId === groupId;

  const unreadCount = feeds.reduce(
    (sum, feed) => sum + (feed.unread_count || 0),
    0,
  );
  const totalCount = feeds.reduce((sum, feed) => sum + (feed.item_count || 0), 0);
  const displayCount =
    unreadCount > 0 ? unreadCount : showTotalCount || isSelected ? totalCount : 0;

  return (
    <Collapsible
      open={isOpen}
      onOpenChange={setIsOpen}
      className="w-full min-w-0"
    >
      <div
        className={cn(
          "sidebar-row flex h-[30px] w-full min-w-0 items-center gap-1.5 rounded-md px-2 text-[13px] transition-colors",
        )}
        data-selected={isSelected}
      >
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            setIsOpen(!isOpen);
          }}
          className="-m-1 shrink-0 rounded-md p-1 transition-colors hover:bg-foreground/10"
        >
          <ChevronRight
            className={cn(
              "h-4 w-4 text-muted-foreground transition-transform",
              isOpen && "rotate-90",
            )}
          />
        </button>
        <button
          type="button"
          onClick={() => setSelectedGroup(groupId)}
          className="flex min-w-0 flex-1 items-center gap-1.5 text-left"
        >
          <span className="block min-w-0 flex-1 truncate">{name}</span>
          {displayCount > 0 && (
            <span
              className={cn(
                "sidebar-count shrink-0",
                unreadCount === 0 && "text-muted-foreground/70",
              )}
            >
              {displayCount}
            </span>
          )}
        </button>
      </div>
      <CollapsibleContent>
        <div className="w-full min-w-0 pl-[18px] pt-0.5">
          {feeds.map((feed) => (
            <FeedItem key={feed.id} feed={feed} />
          ))}
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}
