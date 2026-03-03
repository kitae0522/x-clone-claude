import { useQuery } from '@tanstack/react-query'
import type { APIResponse, Post } from '@/types/api'

async function fetchPosts(): Promise<Post[]> {
  const res = await fetch('/api/posts')
  if (!res.ok) {
    throw new Error(`Failed to fetch posts: ${res.status}`)
  }
  const json: APIResponse<Post[]> = await res.json()
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
