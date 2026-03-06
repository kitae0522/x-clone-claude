import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type {
  APIResponse,
  FollowListResponse,
  FollowStatusResponse,
} from '@/types/api'
import { apiFetch } from '@/lib/api'

async function postFollow(handle: string): Promise<FollowStatusResponse> {
  const res = await apiFetch(`/api/users/${handle}/follow`, { method: 'POST' })
  const json: APIResponse<FollowStatusResponse> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Failed to follow')
  }
  return json.data
}

async function deleteFollow(handle: string): Promise<FollowStatusResponse> {
  const res = await apiFetch(`/api/users/${handle}/follow`, { method: 'DELETE' })
  const json: APIResponse<FollowStatusResponse> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Failed to unfollow')
  }
  return json.data
}

async function fetchFollowing(handle: string): Promise<FollowListResponse> {
  const res = await apiFetch(`/api/users/${handle}/following`)
  const json: APIResponse<FollowListResponse> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Failed to fetch following')
  }
  return json.data
}

async function fetchFollowers(handle: string): Promise<FollowListResponse> {
  const res = await apiFetch(`/api/users/${handle}/followers`)
  const json: APIResponse<FollowListResponse> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Failed to fetch followers')
  }
  return json.data
}

export function useFollow(handle: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => postFollow(handle),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users', handle] })
    },
  })
}

export function useUnfollow(handle: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => deleteFollow(handle),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users', handle] })
    },
  })
}

export function useFollowing(handle: string, enabled: boolean) {
  return useQuery({
    queryKey: ['users', handle, 'following'],
    queryFn: () => fetchFollowing(handle),
    enabled,
  })
}

export function useFollowers(handle: string, enabled: boolean) {
  return useQuery({
    queryKey: ['users', handle, 'followers'],
    queryFn: () => fetchFollowers(handle),
    enabled,
  })
}
