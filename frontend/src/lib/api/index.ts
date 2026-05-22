import { api } from "./client";
import type {
  APIResponse,
  ListAPIResponse,
  LoginRequest,
  Group,
  Feed,
  Item,
  Bookmark,
  ReadLaterItem,
  CreateGroupRequest,
  UpdateGroupRequest,
  CreateFeedRequest,
  UpdateFeedRequest,
  ValidateFeedRequest,
  ValidateFeedResponse,
  CreateBookmarkRequest,
  CreateReadLaterItemRequest,
  MarkItemsReadRequest,
  ListItemsParams,
  BatchCreateFeedsRequest,
  BatchCreateFeedsResponse,
  SearchResponse,
  OIDCStatusResponse,
  OIDCLoginResponse,
  RetentionSettings,
  UpdateRetentionSettingsRequest,
} from "./types";

// Session APIs
export const sessionAPI = {
  login: (data: LoginRequest) =>
    api.post<APIResponse<{ message: string }>>("/sessions", data),

  logout: () => api.delete<void>("/sessions"),
};

// OIDC APIs
export const oidcAPI = {
  status: () => api.get<APIResponse<OIDCStatusResponse>>("/oidc/enabled"),

  login: () => api.get<APIResponse<OIDCLoginResponse>>("/oidc/login"),
};

// Group APIs
export const groupAPI = {
  list: () => api.get<ListAPIResponse<Group>>("/groups"),

  get: (id: number) => api.get<APIResponse<Group>>(`/groups/${id}`),

  create: (data: CreateGroupRequest) =>
    api.post<APIResponse<Group>>("/groups", data),

  update: (id: number, data: UpdateGroupRequest) =>
    api.patch<APIResponse<Group>>(`/groups/${id}`, data),

  delete: (id: number) => api.delete<void>(`/groups/${id}`),
};

// Feed APIs
export const feedAPI = {
  list: () => api.get<ListAPIResponse<Feed>>("/feeds"),

  get: (id: number) => api.get<APIResponse<Feed>>(`/feeds/${id}`),

  create: (data: CreateFeedRequest) =>
    api.post<APIResponse<Feed>>("/feeds", data),

  update: (id: number, data: UpdateFeedRequest) =>
    api.patch<APIResponse<Feed>>(`/feeds/${id}`, data),

  delete: (id: number) => api.delete<void>(`/feeds/${id}`),

  validate: (data: ValidateFeedRequest) =>
    api.post<APIResponse<ValidateFeedResponse>>("/feeds/validate", data),

  refresh: () => api.post<void>("/feeds/refresh"),

  refreshOne: (id: number) => api.post<void>(`/feeds/${id}/refresh`),

  batchCreate: (data: BatchCreateFeedsRequest) =>
    api.post<APIResponse<BatchCreateFeedsResponse>>("/feeds/batch", data),
};

// Item APIs
export const itemAPI = {
  list: (params?: ListItemsParams) => {
    const query = new URLSearchParams();
    if (params?.feed_id) query.set("feed_id", params.feed_id.toString());
    if (params?.group_id) query.set("group_id", params.group_id.toString());
    if (params?.unread !== undefined)
      query.set("unread", params.unread.toString());
    if (params?.limit) query.set("limit", params.limit.toString());
    if (params?.offset) query.set("offset", params.offset.toString());
    if (params?.order_by) query.set("order_by", params.order_by);

    const queryString = query.toString();
    return api.get<ListAPIResponse<Item>>(
      `/items${queryString ? `?${queryString}` : ""}`,
    );
  },

  get: (id: number) => api.get<APIResponse<Item>>(`/items/${id}`),

  markRead: (data: MarkItemsReadRequest) =>
    api.patch<void>("/items/-/read", data),

  markUnread: (data: MarkItemsReadRequest) =>
    api.patch<void>("/items/-/unread", data),
};

// Bookmark APIs
export const bookmarkAPI = {
  list: (limit = 50, offset = 0) => {
    const query = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });
    return api.get<ListAPIResponse<Bookmark>>(`/bookmarks?${query}`);
  },

  get: (id: number) => api.get<APIResponse<Bookmark>>(`/bookmarks/${id}`),

  create: (data: CreateBookmarkRequest) =>
    api.post<APIResponse<Bookmark>>("/bookmarks", data),

  delete: (id: number) => api.delete<void>(`/bookmarks/${id}`),
};

// Read Later APIs
export const readLaterAPI = {
  list: (limit = 50, offset = 0) => {
    const query = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });
    return api.get<ListAPIResponse<ReadLaterItem>>(`/read-later?${query}`);
  },

  get: (id: number) =>
    api.get<APIResponse<ReadLaterItem>>(`/read-later/${id}`),

  create: (data: CreateReadLaterItemRequest) =>
    api.post<APIResponse<ReadLaterItem>>("/read-later", data),

  delete: (id: number) => api.delete<void>(`/read-later/${id}`),
};

// Search APIs
export const searchAPI = {
  search: (q: string, limit = 10) =>
    api.get<APIResponse<SearchResponse>>(
      `/search?q=${encodeURIComponent(q)}&limit=${limit}`,
    ),
};

export const settingsAPI = {
  getRetention: () =>
    api.get<APIResponse<RetentionSettings>>("/settings/retention"),

  updateRetention: (data: UpdateRetentionSettingsRequest) =>
    api.patch<APIResponse<RetentionSettings>>("/settings/retention", data),
};

export * from "./types";
export { APIError, setUnauthorizedCallback } from "./client";
