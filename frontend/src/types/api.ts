export interface Post {
  id: string
  authorId: string
  content: string
  visibility: 'public' | 'friends' | 'private'
  createdAt: string
  updatedAt: string
}

export interface APIResponse<T> {
  success: boolean
  data: T
  error: string | null
}

export interface User {
  id: string
  email: string
  username: string
  displayName: string
  bio: string
  profileImageUrl: string
  headerImageUrl: string
  createdAt: string
  updatedAt: string
}

export interface ProfileUser {
  id: string
  username: string
  displayName: string
  bio: string
  profileImageUrl: string
  headerImageUrl: string
  followersCount: number
  followingCount: number
  isFollowing: boolean
  createdAt: string
  updatedAt: string
}

export interface FollowUser {
  id: string
  username: string
  displayName: string
  bio: string
  profileImageUrl: string
}

export interface FollowListResponse {
  users: FollowUser[]
  total: number
}

export interface FollowStatusResponse {
  following: boolean
}

export interface RegisterRequest {
  email: string
  username: string
  password: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface UpdateProfileRequest {
  displayName: string
  bio: string
  username: string
  profileImageUrl: string
  headerImageUrl: string
}

export interface CreatePostRequest {
  content: string
  visibility: 'public' | 'friends' | 'private'
}

export interface PostAuthor {
  username: string
  displayName: string
  profileImageUrl: string
}

export interface PostDetail {
  id: string
  authorId: string
  parentId: string | null
  content: string
  visibility: 'public' | 'friends' | 'private'
  author: PostAuthor
  likeCount: number
  replyCount: number
  isLiked: boolean
  topReplies: PostDetail[] | null
  createdAt: string
  updatedAt: string
}

export interface CreateReplyRequest {
  content: string
}

export interface LikeStatusResponse {
  liked: boolean
}
