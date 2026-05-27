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
        "article-row group relative flex w-full cursor-pointer items-start gap-2 rounded-lg border px-2.5 py-2 text-left transition-all",
        compact && "px-2 py-2",
      )}
      data-selected={isSelected}
      data-unread={article.unread}
    >
      <span
        className={cn(
          "mt-[1.02rem] size-1.5 shrink-0 rounded-full transition-all",
          article.unread
            ? "bg-primary shadow-[0_0_0_2px_oklch(0.62_0.18_254_/_13%)]"
            : "bg-transparent",
        )}
        aria-hidden="true"
      />
      {/* Article Content */}
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <div className="flex items-center gap-1.5 text-[11px] leading-none text-muted-foreground">
          <FeedFavicon src={feedFaviconUrl} className="h-4 w-4 rounded-sm" />
          <span className="min-w-0 truncate font-medium">{feedName}</span>
          <span className="text-muted-foreground/60">·</span>
          <span className="shrink-0">{formatDate(article.pub_date)}</span>
        </div>
        <h3
          className={cn(
            "line-clamp-2 text-[14px] leading-snug",
            article.unread
              ? "font-semibold text-foreground"
              : "font-medium text-foreground/66",
          )}
        >
          {article.title}
        </h3>
        <p
          className={cn(
            "line-clamp-2 text-[12.5px] leading-relaxed text-muted-foreground/78",
            compact && "line-clamp-1",
          )}
        >
          {extractSummary(article.content, 150)}
        </p>
      </div>
    </div>
  );
}

export const ArticleItem = memo(ArticleItemComponent);
