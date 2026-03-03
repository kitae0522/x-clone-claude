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
  createdAt: string
  updatedAt: string
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
