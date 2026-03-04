import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { APIResponse, PostDetail, CreateReplyRequest } from '@/types/api'

async function fetchReplies(postId: string): Promise<PostDetail[]> {
  const res = await fetch(`/api/posts/${postId}/replies`)
  if (!res.ok) {
    throw new Error(`Failed to fetch replies: ${res.status}`)
  }
  const json: APIResponse<PostDetail[]> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Unknown error')
  }
  return json.data
}

async function createReply(postId: string, req: CreateReplyRequest): Promise<PostDetail> {
  const res = await fetch(`/api/posts/${postId}/reply`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  if (!res.ok) {
    const json = await res.json()
    throw new Error(json.error ?? `Failed to create reply: ${res.status}`)
  }
  const json: APIResponse<PostDetail> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Unknown error')
  }
  return json.data
}

export function useReplies(postId: string) {
  return useQuery({
    queryKey: ['post', postId, 'replies'],
    queryFn: () => fetchReplies(postId),
    enabled: !!postId,
  })
}

export function useCreateReply(postId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (req: CreateReplyRequest) => createReply(postId, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['post', postId, 'replies'] })
      queryClient.invalidateQueries({ queryKey: ['post', postId] })
      queryClient.invalidateQueries({ queryKey: ['posts'] })
    },
  })
}
