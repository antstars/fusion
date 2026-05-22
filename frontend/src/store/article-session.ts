import { create } from "zustand";

interface ArticleSessionState {
  starredOverrides: Record<number, boolean>;
  readLaterOverrides: Record<number, boolean>;
  setStarredOverride: (itemId: number, starred: boolean) => void;
  clearStarredOverride: (itemId: number) => void;
  setReadLaterOverride: (itemId: number, readLater: boolean) => void;
  clearReadLaterOverride: (itemId: number) => void;
}

export const useArticleSessionStore = create<ArticleSessionState>((set) => ({
  starredOverrides: {},
  readLaterOverrides: {},
  setStarredOverride: (itemId, starred) =>
    set((state) => ({
      starredOverrides: {
        ...state.starredOverrides,
        [itemId]: starred,
      },
    })),
  clearStarredOverride: (itemId) =>
    set((state) => {
      const next = { ...state.starredOverrides };
      delete next[itemId];
      return { starredOverrides: next };
    }),
  setReadLaterOverride: (itemId, readLater) =>
    set((state) => ({
      readLaterOverrides: {
        ...state.readLaterOverrides,
        [itemId]: readLater,
      },
    })),
  clearReadLaterOverride: (itemId) =>
    set((state) => {
      const next = { ...state.readLaterOverrides };
      delete next[itemId];
      return { readLaterOverrides: next };
    }),
}));
