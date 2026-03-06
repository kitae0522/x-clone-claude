import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { APIResponse, LikeStatusResponse, PostDetail } from "@/types/api";
import { apiFetch } from "@/lib/api";

async function postLike(postId: string): Promise<LikeStatusResponse> {
  const res = await apiFetch(`/api/posts/${postId}/like`, { method: "POST" });
  const json: APIResponse<LikeStatusResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to like");
  }
  return json.data;
}

async function deleteLike(postId: string): Promise<LikeStatusResponse> {
  const res = await apiFetch(`/api/posts/${postId}/like`, { method: "DELETE" });
  const json: APIResponse<LikeStatusResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to unlike");
  }
  return json.data;
}

function updatePostInCache(old: PostDetail, liked: boolean): PostDetail {
  return {
    ...old,
    isLiked: liked,
    likeCount: old.likeCount + (liked ? 1 : -1),
  };
}

export function useLike(postId: string, isLiked: boolean, parentId?: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => (isLiked ? deleteLike(postId) : postLike(postId)),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: ["posts"] });
      await queryClient.cancelQueries({ queryKey: ["post", postId] });
      if (parentId) {
        await queryClient.cancelQueries({
          queryKey: ["post", parentId, "replies"],
        });
      }

      const prevPosts = queryClient.getQueryData<PostDetail[]>(["posts"]);
      const prevPost = queryClient.getQueryData<PostDetail>(["post", postId]);
      const prevReplies = parentId
        ? queryClient.getQueryData<PostDetail[]>(["post", parentId, "replies"])
        : undefined;

      const nextLiked = !isLiked;

      if (prevPosts) {
        queryClient.setQueryData<PostDetail[]>(["posts"], (old) =>
          old?.map((p) =>
            p.id === postId ? updatePostInCache(p, nextLiked) : p,
          ),
        );
      }

      if (prevPost) {
        queryClient.setQueryData<PostDetail>(["post", postId], (old) =>
          old ? updatePostInCache(old, nextLiked) : old,
        );
      }

      if (parentId && prevReplies) {
        queryClient.setQueryData<PostDetail[]>(
          ["post", parentId, "replies"],
          (old) =>
            old?.map((r) =>
              r.id === postId ? updatePostInCache(r, nextLiked) : r,
            ),
        );
      }

      return { prevPosts, prevPost, prevReplies };
    },
    onError: (_err, _vars, context) => {
      if (context?.prevPosts) {
        queryClient.setQueryData(["posts"], context.prevPosts);
      }
      if (context?.prevPost) {
        queryClient.setQueryData(["post", postId], context.prevPost);
      }
      if (parentId && context?.prevReplies) {
        queryClient.setQueryData(
          ["post", parentId, "replies"],
          context.prevReplies,
        );
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["posts"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      if (parentId) {
        queryClient.invalidateQueries({
          queryKey: ["post", parentId, "replies"],
        });
      }
    },
  });
}
