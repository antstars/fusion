import { create } from "zustand";

interface ArticleSessionState {
  unreadOverrides: Record<number, boolean>;
  starredOverrides: Record<number, boolean>;
  readLaterOverrides: Record<number, boolean>;
  setUnreadOverride: (itemId: number, unread: boolean) => void;
  clearUnreadOverride: (itemId: number) => void;
  clearUnreadOverrides: (itemIds: number[]) => void;
  setStarredOverride: (itemId: number, starred: boolean) => void;
  clearStarredOverride: (itemId: number) => void;
  setReadLaterOverride: (itemId: number, readLater: boolean) => void;
  clearReadLaterOverride: (itemId: number) => void;
}

export const useArticleSessionStore = create<ArticleSessionState>((set) => ({
  unreadOverrides: {},
  starredOverrides: {},
  readLaterOverrides: {},
  setUnreadOverride: (itemId, unread) =>
    set((state) => ({
      unreadOverrides: {
        ...state.unreadOverrides,
        [itemId]: unread,
      },
    })),
  clearUnreadOverride: (itemId) =>
    set((state) => {
      const next = { ...state.unreadOverrides };
      delete next[itemId];
      return { unreadOverrides: next };
    }),
  clearUnreadOverrides: (itemIds) =>
    set((state) => {
      const next = { ...state.unreadOverrides };
      for (const itemId of itemIds) {
        delete next[itemId];
      }
      return { unreadOverrides: next };
    }),
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
