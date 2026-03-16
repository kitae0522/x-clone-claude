import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type {
  APIResponse,
  TrashListResponse,
  RestorePostResponse,
  PermanentDeleteResponse,
} from "@/types/api";
import { apiFetch } from "@/lib/api";
import { toast } from "sonner";

async function fetchTrash(cursor?: string): Promise<TrashListResponse> {
  const params = new URLSearchParams();
  if (cursor) params.set("cursor", cursor);
  const url = `/api/users/trash${params.toString() ? `?${params}` : ""}`;
  const res = await apiFetch(url);
  const json: APIResponse<TrashListResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to fetch trash");
  }
  return json.data;
}

async function restorePost(postId: string): Promise<RestorePostResponse> {
  const res = await apiFetch(`/api/posts/${postId}/restore`, { method: "PUT" });
  const json: APIResponse<RestorePostResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to restore post");
  }
  return json.data;
}

async function permanentDeletePost(postId: string): Promise<PermanentDeleteResponse> {
  const res = await apiFetch(`/api/posts/${postId}/permanent`, { method: "DELETE" });
  const json: APIResponse<PermanentDeleteResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to permanently delete post");
  }
  return json.data;
}

export function useTrash() {
  return useInfiniteQuery({
    queryKey: ["trash"],
    queryFn: ({ pageParam }) => fetchTrash(pageParam),
    getNextPageParam: (lastPage) => lastPage.hasMore ? (lastPage.nextCursor ?? undefined) : undefined,
    initialPageParam: undefined as string | undefined,
  });
}

export function useRestorePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: restorePost,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["trash"] });
      queryClient.invalidateQueries({ queryKey: ["posts"] });
      toast.success("게시글이 복원되었습니다");
    },
    onError: (err) => {
      toast.error(err.message);
    },
  });
}

export function usePermanentDelete() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: permanentDeletePost,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["trash"] });
      toast.success("게시글이 영구 삭제되었습니다");
    },
    onError: (err) => {
      toast.error(err.message);
    },
  });
}
