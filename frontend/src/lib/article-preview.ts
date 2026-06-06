import type { Item } from "@/lib/api";
import { extractFirstImageUrl, extractSummary } from "@/lib/utils";

interface ArticlePreview {
  summary: string;
  thumbnailUrl: string | null;
}

const maxPreviewCacheEntries = 1000;
const previewCache = new Map<string, ArticlePreview>();

function getPreviewCacheKey(article: Item): string {
  return `${article.id}\0${article.link}\0${article.content}`;
}

function rememberArticlePreview(key: string, preview: ArticlePreview) {
  if (previewCache.size >= maxPreviewCacheEntries) {
    const oldestKey = previewCache.keys().next().value;
    if (oldestKey !== undefined) {
      previewCache.delete(oldestKey);
    }
  }

  previewCache.set(key, preview);
}

export function getArticlePreview(article: Item): ArticlePreview {
  const key = getPreviewCacheKey(article);
  const cached = previewCache.get(key);
  if (cached) return cached;

  const preview = {
    summary: extractSummary(article.content, 150),
    thumbnailUrl: extractFirstImageUrl(article.content, article.link),
  };
  rememberArticlePreview(key, preview);
  return preview;
}
