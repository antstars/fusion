import { memo } from "react";
import { Circle, CircleCheck, Star, ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/lib/i18n";
import { cn, formatDate, extractSummary } from "@/lib/utils";
import type { Item } from "@/lib/api";
import { FeedFavicon } from "@/components/feed/feed-favicon";
import { toSafeExternalUrl } from "@/lib/safe-url";

interface ArticleItemProps {
  article: Item;
  isSelected: boolean;
  onSelectArticle: (articleId: number | null) => void;
  onToggleRead: (article: Item) => Promise<void>;
  onToggleStar: (article: Item) => Promise<void>;
  canToggleRead: boolean;
  isStarred: boolean;
  feedName: string;
  feedFaviconUrl: string | null;
  compact?: boolean;
}

function ArticleItemComponent({
  article,
  isSelected,
  onSelectArticle,
  onToggleRead,
  onToggleStar,
  canToggleRead,
  isStarred,
  feedName,
  feedFaviconUrl,
  compact = false,
}: ArticleItemProps) {
  const { t } = useI18n();

  const safeArticleLink = toSafeExternalUrl(article.link);

  const handleToggleRead = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!canToggleRead) return;

    try {
      await onToggleRead(article);
    } catch (error) {
      console.error("Failed to toggle read status:", error);
    }
  };

  const handleToggleStar = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await onToggleStar(article);
    } catch (error) {
      console.error("Failed to toggle star:", error);
    }
  };

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

      {/* Article Actions */}
      <div className="absolute right-2 top-2 hidden items-center gap-1 group-hover:flex">
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={handleToggleRead}
          disabled={!canToggleRead}
          className={cn(
            "border bg-background shadow-xs",
            article.unread ? "border-border" : "border-primary/20 bg-primary/10",
          )}
          aria-label={
            article.unread
              ? t("article.action.markRead")
              : t("article.action.markUnread")
          }
          title={
            article.unread
              ? t("article.action.markRead")
              : t("article.action.markUnread")
          }
        >
          {article.unread ? (
            <Circle className="text-muted-foreground" />
          ) : (
            <CircleCheck className="text-primary" />
          )}
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={handleToggleStar}
          className={cn(
            "border bg-background shadow-xs",
            isStarred ? "bg-amber-100 dark:bg-amber-950/40" : "bg-muted",
          )}
          aria-label={
            isStarred ? t("article.action.unstar") : t("article.action.star")
          }
          title={isStarred ? t("article.action.unstar") : t("article.action.star")}
        >
          <Star
            className={cn(
              isStarred
                ? "fill-amber-500 text-amber-500"
                : "text-muted-foreground",
            )}
          />
        </Button>
        {safeArticleLink ? (
          <Button
            asChild
            variant="ghost"
            size="icon-sm"
            className="border bg-background shadow-xs"
            aria-label={t("article.action.openInBrowser")}
            title={t("article.action.openInBrowser")}
          >
            <a
              href={safeArticleLink}
              target="_blank"
              rel="noopener noreferrer"
              onClick={(e) => e.stopPropagation()}
            >
              <ExternalLink className="text-muted-foreground" />
            </a>
          </Button>
        ) : (
          <Button
            variant="ghost"
            size="icon-sm"
            disabled
            className="border bg-background shadow-xs"
            aria-label={t("article.action.openInBrowser")}
            title={t("article.action.openInBrowser")}
          >
            <ExternalLink className="text-muted-foreground" />
          </Button>
        )}
      </div>
    </div>
  );
}

export const ArticleItem = memo(ArticleItemComponent);
