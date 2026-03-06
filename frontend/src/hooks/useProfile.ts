import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { APIResponse, ProfileUser, UpdateProfileRequest, User } from '@/types/api'
import { apiFetch } from '@/lib/api'

async function fetchProfile(handle: string): Promise<ProfileUser> {
  const res = await apiFetch(`/api/users/${handle}`)
  const json: APIResponse<ProfileUser> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Failed to fetch profile')
  }
  return json.data
}

async function putUpdateProfile(data: UpdateProfileRequest): Promise<User> {
  const res = await apiFetch('/api/users/profile', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  const json: APIResponse<User> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Failed to update profile')
  }
  return json.data
}

export function useProfile(handle: string, enabled: boolean = true) {
  return useQuery({
    queryKey: ['users', handle],
    queryFn: () => fetchProfile(handle),
    enabled: !!handle && enabled,
  })
}

export function useUpdateProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: putUpdateProfile,
    onSuccess: (user) => {
      queryClient.setQueryData(['auth', 'me'], user)
      queryClient.invalidateQueries({ queryKey: ['users', user.username] })
    },
  })
}
