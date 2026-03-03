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
