import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { APIResponse, PostDetail, CreatePostRequest } from '@/types/api'

async function fetchPosts(): Promise<PostDetail[]> {
  const res = await fetch('/api/posts')
  if (!res.ok) {
    throw new Error(`Failed to fetch posts: ${res.status}`)
  }
  const json: APIResponse<PostDetail[]> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Unknown error')
  }
  return json.data
}

async function fetchPostDetail(id: string): Promise<PostDetail> {
  const res = await fetch(`/api/posts/${id}`)
  if (!res.ok) {
    throw new Error(`Failed to fetch post: ${res.status}`)
  }
  const json: APIResponse<PostDetail> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Unknown error')
  }
  return json.data
}

async function createPost(req: CreatePostRequest): Promise<PostDetail> {
  const res = await fetch('/api/posts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  if (!res.ok) {
    const json = await res.json()
    throw new Error(json.error ?? `Failed to create post: ${res.status}`)
  }
  const json: APIResponse<PostDetail> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Unknown error')
  }
  return json.data
}

export function usePosts() {
  return useQuery({
    queryKey: ['posts'],
    queryFn: fetchPosts,
  })
}

export function usePostDetail(id: string) {
  return useQuery({
    queryKey: ['post', id],
    queryFn: () => fetchPostDetail(id),
    enabled: !!id,
  })
}

export function useCreatePost() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createPost,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['posts'] })
    },
  })
}
