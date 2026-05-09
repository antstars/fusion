import { memo } from "react";
import { cn, formatDate, extractSummary } from "@/lib/utils";
import type { Item } from "@/lib/api";
import { FeedFavicon } from "@/components/feed/feed-favicon";

interface ArticleItemProps {
  article: Item;
  isSelected: boolean;
  onSelectArticle: (articleId: number | null) => void;
  feedName: string;
  feedFaviconUrl: string | null;
  compact?: boolean;
}

function ArticleItemComponent({
  article,
  isSelected,
  onSelectArticle,
  feedName,
  feedFaviconUrl,
  compact = false,
}: ArticleItemProps) {
  return (
    <div
      role="button"
      tabIndex={0}
      onClick={() => onSelectArticle(article.id)}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelectArticle(article.id);
        }
      }}
      className={cn(
        "article-row group relative flex w-full cursor-pointer items-start gap-4 rounded-md border px-4 py-4 text-left transition-colors",
        compact && "px-3 py-3",
      )}
      data-selected={isSelected}
    >
      {/* Article Content */}
      <div className="flex min-w-0 flex-1 flex-col gap-1.5">
        <h3
          className={cn(
            "line-clamp-2 text-[15px] leading-snug",
            article.unread
              ? "font-semibold text-foreground"
              : "font-medium text-muted-foreground",
          )}
        >
          {article.title}
        </h3>
        <p
          className={cn(
            "line-clamp-2 text-sm leading-relaxed text-muted-foreground",
            compact && "line-clamp-1 text-xs",
          )}
        >
          {extractSummary(article.content, 150)}
        </p>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <FeedFavicon src={feedFaviconUrl} className="h-3.5 w-3.5 rounded-sm" />
          <span className="truncate font-medium">
            {feedName}
          </span>
          <span>-</span>
          <span className="shrink-0">
            {formatDate(article.pub_date)}
          </span>
        </div>
      </div>
    </div>
  );
}

export const ArticleItem = memo(ArticleItemComponent);
