import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { APIResponse, LoginRequest, RegisterRequest, User } from '@/types/api'

async function fetchMe(): Promise<User> {
  const res = await fetch('/api/auth/me')
  if (!res.ok) {
    throw new Error('Not authenticated')
  }
  const json: APIResponse<User> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Unknown error')
  }
  return json.data
}

async function postRegister(data: RegisterRequest): Promise<User> {
  const res = await fetch('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  const json: APIResponse<User> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Registration failed')
  }
  return json.data
}

async function postLogin(data: LoginRequest): Promise<User> {
  const res = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  const json: APIResponse<User> = await res.json()
  if (!json.success) {
    throw new Error(json.error ?? 'Login failed')
  }
  return json.data
}

async function postLogout(): Promise<void> {
  const res = await fetch('/api/auth/logout', { method: 'POST' })
  if (!res.ok) {
    throw new Error('Logout failed')
  }
}

export function useMe() {
  return useQuery({
    queryKey: ['auth', 'me'],
    queryFn: fetchMe,
    retry: false,
  })
}

export function useRegister() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: postRegister,
    onSuccess: (user) => {
      queryClient.setQueryData(['auth', 'me'], user)
    },
  })
}

export function useLogin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: postLogin,
    onSuccess: (user) => {
      queryClient.setQueryData(['auth', 'me'], user)
    },
  })
}

export function useLogout() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: postLogout,
    onSuccess: () => {
      queryClient.setQueryData(['auth', 'me'], null)
      queryClient.invalidateQueries({ queryKey: ['auth', 'me'] })
    },
  })
}
