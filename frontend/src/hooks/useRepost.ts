import { useMutation, useQueryClient } from "@tanstack/react-query";
import type {
  APIResponse,
  RepostStatusResponse,
  PostDetail,
} from "@/types/api";
import { apiFetch } from "@/lib/api";

async function postRepost(postId: string): Promise<RepostStatusResponse> {
  const res = await apiFetch(`/api/posts/${postId}/repost`, { method: "POST" });
  const json: APIResponse<RepostStatusResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to repost");
  }
  return json.data;
}

async function deleteRepost(postId: string): Promise<RepostStatusResponse> {
  const res = await apiFetch(`/api/posts/${postId}/repost`, {
    method: "DELETE",
  });
  const json: APIResponse<RepostStatusResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to unrepost");
  }
  return json.data;
}

function updatePostInCache(old: PostDetail, reposted: boolean): PostDetail {
  return {
    ...old,
    isReposted: reposted,
    repostCount: old.repostCount + (reposted ? 1 : -1),
  };
}

export function useRepost(
  postId: string,
  isReposted: boolean,
  parentId?: string,
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => (isReposted ? deleteRepost(postId) : postRepost(postId)),
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

      const nextReposted = !isReposted;

      if (prevPosts) {
        queryClient.setQueryData<PostDetail[]>(["posts"], (old) =>
          old?.map((p) =>
            p.id === postId ? updatePostInCache(p, nextReposted) : p,
          ),
        );
      }

      if (prevPost) {
        queryClient.setQueryData<PostDetail>(["post", postId], (old) =>
          old ? updatePostInCache(old, nextReposted) : old,
        );
      }

      if (parentId && prevReplies) {
        queryClient.setQueryData<PostDetail[]>(
          ["post", parentId, "replies"],
          (old) =>
            old?.map((r) =>
              r.id === postId ? updatePostInCache(r, nextReposted) : r,
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
