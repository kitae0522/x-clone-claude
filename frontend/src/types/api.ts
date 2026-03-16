export interface Post {
  id: string;
  authorId: string;
  content: string;
  visibility: "public" | "follower" | "private";
  createdAt: string;
  updatedAt: string;
}

export interface APIResponse<T> {
  success: boolean;
  data: T;
  error: string | null;
}

export interface User {
  id: string;
  email: string;
  username: string;
  displayName: string;
  bio: string;
  profileImageUrl: string;
  headerImageUrl: string;
  createdAt: string;
  updatedAt: string;
}

export interface ProfileUser {
  id: string;
  username: string;
  displayName: string;
  bio: string;
  profileImageUrl: string;
  headerImageUrl: string;
  followersCount: number;
  followingCount: number;
  isFollowing: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface FollowUser {
  id: string;
  username: string;
  displayName: string;
  bio: string;
  profileImageUrl: string;
}

export interface FollowListResponse {
  users: FollowUser[];
  total: number;
}

export interface FollowStatusResponse {
  following: boolean;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface UpdateProfileRequest {
  displayName: string;
  bio: string;
  username: string;
  profileImageUrl: string;
  headerImageUrl: string;
}

export interface MediaItem {
  id: string;
  url: string;
  type: "image" | "video" | "gif";
  mimeType: string;
  width: number | null;
  height: number | null;
  size: number;
  duration: number | null;
  status?: "pending" | "processing" | "ready" | "failed";
}

export interface LocationData {
  latitude: number;
  longitude: number;
  name: string;
}

export interface PollOption {
  text: string;
  voteCount: number;
}

export interface PollData {
  options: PollOption[];
  totalVotes: number;
  votedIndex: number;
  expiresAt: string;
  isExpired: boolean;
}

export interface CreatePostRequest {
  content: string;
  visibility: "public" | "follower" | "private";
  mediaIds?: string[];
  location?: {
    latitude: number;
    longitude: number;
    name?: string;
  };
  poll?: {
    options: string[];
    durationMinutes: number;
  };
}

export interface PostAuthor {
  username: string;
  displayName: string;
  profileImageUrl: string;
  isDeleted?: boolean;
}

export interface ParentPostSummary {
  id: string;
  content: string;
  author: PostAuthor;
}

export interface PostDetail {
  id: string;
  authorId: string;
  parentId: string | null;
  parent?: ParentPostSummary | null;
  content: string;
  visibility: "public" | "follower" | "private";
  author: PostAuthor;
  likeCount: number;
  replyCount: number;
  viewCount: number;
  repostCount: number;
  isLiked: boolean;
  isBookmarked: boolean;
  isReposted: boolean;
  repostedBy?: { username: string; displayName: string } | null;
  media?: MediaItem[] | null;
  location?: LocationData | null;
  poll?: PollData | null;
  topReplies: PostDetail[] | null;
  createdAt: string;
  updatedAt: string;
}

export interface UpdatePostRequest {
  content?: string;
  visibility?: "public" | "follower" | "private";
  mediaIds?: string[];
  location?: {
    latitude: number;
    longitude: number;
    name?: string;
  } | null;
  clearLocation?: boolean;
  poll?: {
    options: string[];
    durationMinutes: number;
  } | null;
  clearPoll?: boolean;
}

export interface CreateReplyRequest {
  content: string;
  mediaIds?: string[];
  location?: {
    latitude: number;
    longitude: number;
    name?: string;
  };
  poll?: {
    options: string[];
    durationMinutes: number;
  };
}

export interface LikeStatusResponse {
  liked: boolean;
}

export interface RepostStatusResponse {
  reposted: boolean;
}

export interface BookmarkStatusResponse {
  bookmarked: boolean;
}

export interface BookmarkListResponse {
  posts: PostDetail[];
  nextCursor: string;
  hasMore: boolean;
}
