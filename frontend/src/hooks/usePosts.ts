import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type {
  APIResponse,
  PostDetail,
  CreatePostRequest,
  UpdatePostRequest,
} from "@/types/api";
import { apiFetch } from "@/lib/api";

async function fetchPosts(): Promise<PostDetail[]> {
  const res = await apiFetch("/api/posts");
  if (!res.ok) {
    throw new Error(`Failed to fetch posts: ${res.status}`);
  }
  const json: APIResponse<PostDetail[]> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Unknown error");
  }
  return json.data;
}

async function fetchPostDetail(id: string): Promise<PostDetail> {
  const res = await apiFetch(`/api/posts/${id}`);
  if (!res.ok) {
    throw new Error(`Failed to fetch post: ${res.status}`);
  }
  const json: APIResponse<PostDetail> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Unknown error");
  }
  return json.data;
}

async function createPost(req: CreatePostRequest): Promise<PostDetail> {
  const body: Record<string, unknown> = {
    content: req.content,
    visibility: req.visibility,
  };
  if (req.mediaIds && req.mediaIds.length > 0) body.mediaIds = req.mediaIds;
  if (req.location) body.location = req.location;
  if (req.poll) body.poll = req.poll;

  const res = await apiFetch("/api/posts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const json = await res.json();
    throw new Error(json.error ?? `Failed to create post: ${res.status}`);
  }
  const json: APIResponse<PostDetail> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Unknown error");
  }
  return json.data;
}

export function usePosts() {
  return useQuery({
    queryKey: ["posts"],
    queryFn: fetchPosts,
  });
}

export function usePostDetail(id: string) {
  return useQuery({
    queryKey: ["post", id],
    queryFn: () => fetchPostDetail(id),
    enabled: !!id,
  });
}

const MAX_PARENT_DEPTH = 10;

async function fetchParentChain(parentId: string): Promise<PostDetail[]> {
  const chain: PostDetail[] = [];
  let currentId: string | null = parentId;
  let depth = 0;

  while (currentId && depth++ < MAX_PARENT_DEPTH) {
    const post = await fetchPostDetail(currentId);
    chain.unshift(post);
    currentId = post.parentId;
  }

  return chain;
}

export function useParentChain(parentId: string | null) {
  return useQuery({
    queryKey: ["post", parentId, "parentChain"],
    queryFn: () => fetchParentChain(parentId!),
    enabled: !!parentId,
  });
}

export function useCreatePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createPost,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["posts"] });
    },
  });
}

export function useUpdatePost(postId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: UpdatePostRequest) => {
      const res = await apiFetch(`/api/posts/${postId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });
      if (!res.ok) {
        const json = await res.json();
        throw new Error(json.error ?? `Failed to update post: ${res.status}`);
      }
      const json: APIResponse<PostDetail> = await res.json();
      return json.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["posts"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
    },
  });
}

export function useDeletePost(postId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const res = await apiFetch(`/api/posts/${postId}`, {
        method: "DELETE",
      });
      if (!res.ok) {
        const json = await res.json();
        throw new Error(json.error ?? `Failed to delete post: ${res.status}`);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["posts"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
    },
  });
}
