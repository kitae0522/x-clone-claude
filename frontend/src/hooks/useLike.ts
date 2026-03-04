import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { APIResponse, LikeStatusResponse, PostDetail } from '@/types/api'

async function postLike(postId: string): Promise<LikeStatusResponse> {
  const res = await fetch(`/api/posts/${postId}/like`, { method: 'POST' })
  const json: APIResponse<LikeStatusResponse> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Failed to like')
  }
  return json.data
}

async function deleteLike(postId: string): Promise<LikeStatusResponse> {
  const res = await fetch(`/api/posts/${postId}/like`, { method: 'DELETE' })
  const json: APIResponse<LikeStatusResponse> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Failed to unlike')
  }
  return json.data
}

function updatePostInCache(old: PostDetail, liked: boolean): PostDetail {
  return {
    ...old,
    isLiked: liked,
    likeCount: old.likeCount + (liked ? 1 : -1),
  }
}

export function useLike(postId: string, isLiked: boolean) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => (isLiked ? deleteLike(postId) : postLike(postId)),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: ['posts'] })
      await queryClient.cancelQueries({ queryKey: ['post', postId] })

      const prevPosts = queryClient.getQueryData<PostDetail[]>(['posts'])
      const prevPost = queryClient.getQueryData<PostDetail>(['post', postId])

      const nextLiked = !isLiked

      if (prevPosts) {
        queryClient.setQueryData<PostDetail[]>(['posts'], (old) =>
          old?.map((p) =>
            p.id === postId ? updatePostInCache(p, nextLiked) : p,
          ),
        )
      }

      if (prevPost) {
        queryClient.setQueryData<PostDetail>(['post', postId], (old) =>
          old ? updatePostInCache(old, nextLiked) : old,
        )
      }

      return { prevPosts, prevPost }
    },
    onError: (_err, _vars, context) => {
      if (context?.prevPosts) {
        queryClient.setQueryData(['posts'], context.prevPosts)
      }
      if (context?.prevPost) {
        queryClient.setQueryData(['post', postId], context.prevPost)
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['posts'] })
      queryClient.invalidateQueries({ queryKey: ['post', postId] })
    },
  })
}
