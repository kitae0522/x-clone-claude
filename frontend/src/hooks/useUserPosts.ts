import { useQuery } from "@tanstack/react-query";
import type { APIResponse, PostDetail } from "@/types/api";

async function fetchUserPosts(handle: string): Promise<PostDetail[]> {
  const res = await fetch(`/api/users/${handle}/posts`);
  const json: APIResponse<PostDetail[]> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to fetch user posts");
  }
  return json.data;
}

async function fetchUserReplies(handle: string): Promise<PostDetail[]> {
  const res = await fetch(`/api/users/${handle}/replies`);
  const json: APIResponse<PostDetail[]> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to fetch user replies");
  }
  return json.data;
}

async function fetchUserLikes(handle: string): Promise<PostDetail[]> {
  const res = await fetch(`/api/users/${handle}/likes`);
  const json: APIResponse<PostDetail[]> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to fetch user likes");
  }
  return json.data;
}

export function useUserPosts(handle: string) {
  return useQuery({
    queryKey: ["users", handle, "posts"],
    queryFn: () => fetchUserPosts(handle),
    enabled: !!handle,
  });
}

export function useUserReplies(handle: string) {
  return useQuery({
    queryKey: ["users", handle, "replies"],
    queryFn: () => fetchUserReplies(handle),
    enabled: !!handle,
  });
}

export function useUserLikes(handle: string) {
  return useQuery({
    queryKey: ["users", handle, "likes"],
    queryFn: () => fetchUserLikes(handle),
    enabled: !!handle,
  });
}
