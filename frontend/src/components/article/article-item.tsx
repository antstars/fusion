import { memo, useEffect, useMemo, useState } from "react";
import {
  cn,
  extractFirstImageUrl,
  extractSummary,
  formatDate,
} from "@/lib/utils";
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

interface ArticleThumbnailProps {
  src: string;
}

function ArticleThumbnail({ src }: ArticleThumbnailProps) {
  const [hasFailed, setHasFailed] = useState(false);

  useEffect(() => {
    setHasFailed(false);
  }, [src]);

  if (hasFailed) return null;

  return (
    <img
      src={src}
      alt=""
      className="h-20 w-20 shrink-0 rounded-md object-cover"
      loading="lazy"
      decoding="async"
      onError={() => setHasFailed(true)}
    />
  );
}

function ArticleItemComponent({
  article,
  isSelected,
  onSelectArticle,
  feedName,
  feedFaviconUrl,
  compact = false,
}: ArticleItemProps) {
  const summary = useMemo(() => extractSummary(article.content, 150), [
    article.content,
  ]);
  const thumbnailUrl = useMemo(
    () => extractFirstImageUrl(article.content, article.link),
    [article.content, article.link],
  );

  return (
    <button
      type="button"
      onClick={() => onSelectArticle(article.id)}
      className={cn(
        "article-row group relative grid w-full cursor-pointer grid-cols-[8px_minmax(0,1fr)_auto] items-start gap-2 border px-2 py-3 text-left transition-all",
        compact && "px-1.5 py-3",
      )}
      data-selected={isSelected}
      data-unread={article.unread}
    >
      <span
        className={cn(
          "mt-[1.1rem] size-1.5 rounded-full transition-all",
          article.unread ? "bg-orange-500" : "bg-transparent",
        )}
        aria-hidden="true"
      />
      <div className="flex min-w-0 flex-col gap-1">
        <div className="flex items-center gap-1.5 text-[11px] font-semibold leading-none text-muted-foreground">
          <FeedFavicon src={feedFaviconUrl} className="h-4 w-4 rounded" />
          <span className="min-w-0 truncate">{feedName}</span>
          <span className="shrink-0 text-muted-foreground/60">·</span>
          <span className="shrink-0">{formatDate(article.pub_date)}</span>
        </div>
        <h3
          title={article.title}
          className={cn(
            "break-words text-[14px] leading-[1.32] tracking-normal",
            article.unread
              ? "line-clamp-3 font-semibold text-foreground"
              : "line-clamp-2 font-medium text-foreground/75",
          )}
        >
          {article.title}
        </h3>
        <p
          className={cn(
            "line-clamp-2 text-[12.5px] leading-[1.55] text-muted-foreground/82",
            compact && "line-clamp-1",
          )}
        >
          {summary}
        </p>
      </div>
      {thumbnailUrl ? <ArticleThumbnail src={thumbnailUrl} /> : null}
    </button>
  );
}

export const ArticleItem = memo(ArticleItemComponent);
