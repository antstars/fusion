import type { Item } from "@/lib/api";

function normalizeArticleLink(link: string) {
  const trimmed = link.trim();
  if (trimmed === "") return "";

  try {
    const url = new URL(trimmed);
    url.hash = "";
    return url.toString();
  } catch {
    const hashIndex = trimmed.indexOf("#");
    return hashIndex >= 0 ? trimmed.slice(0, hashIndex) : trimmed;
  }
}

function articleIdentity(article: Item) {
  const normalizedLink = normalizeArticleLink(article.link);
  if (normalizedLink !== "") {
    return `feed:${article.feed_id}:link:${normalizedLink}`;
  }

  return `id:${article.id}`;
}

export function dedupeArticlesByIdentity(
  articles: Item[],
  selectedArticleId: number | null = null,
) {
  const indexes = new Map<string, number>();
  const deduped: Item[] = [];

  for (const article of articles) {
    const identity = articleIdentity(article);
    const existingIndex = indexes.get(identity);

    if (existingIndex === undefined) {
      indexes.set(identity, deduped.length);
      deduped.push(article);
      continue;
    }

    if (article.id === selectedArticleId) {
      deduped[existingIndex] = article;
    }
  }

  return deduped;
}
