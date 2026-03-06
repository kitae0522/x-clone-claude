import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type {
  APIResponse,
  BookmarkStatusResponse,
  BookmarkListResponse,
  PostDetail,
} from "@/types/api";

async function postBookmark(
  postId: string,
): Promise<BookmarkStatusResponse> {
  const res = await fetch(`/api/posts/${postId}/bookmark`, { method: "POST" });
  const json: APIResponse<BookmarkStatusResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to bookmark");
  }
  return json.data;
}

async function deleteBookmark(
  postId: string,
): Promise<BookmarkStatusResponse> {
  const res = await fetch(`/api/posts/${postId}/bookmark`, {
    method: "DELETE",
  });
  const json: APIResponse<BookmarkStatusResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to unbookmark");
  }
  return json.data;
}

function toggleBookmarkInPost(post: PostDetail, bookmarked: boolean): PostDetail {
  return { ...post, isBookmarked: bookmarked };
}

export function useBookmark(postId: string, isBookmarked: boolean) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () =>
      isBookmarked ? deleteBookmark(postId) : postBookmark(postId),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: ["posts"] });
      await queryClient.cancelQueries({ queryKey: ["post", postId] });

      const prevPosts = queryClient.getQueryData<PostDetail[]>(["posts"]);
      const prevPost = queryClient.getQueryData<PostDetail>(["post", postId]);

      const next = !isBookmarked;

      if (prevPosts) {
        queryClient.setQueryData<PostDetail[]>(["posts"], (old) =>
          old?.map((p) =>
            p.id === postId ? toggleBookmarkInPost(p, next) : p,
          ),
        );
      }

      if (prevPost) {
        queryClient.setQueryData<PostDetail>(["post", postId], (old) =>
          old ? toggleBookmarkInPost(old, next) : old,
        );
      }

      return { prevPosts, prevPost };
    },
    onError: (_err, _vars, context) => {
      if (context?.prevPosts) {
        queryClient.setQueryData(["posts"], context.prevPosts);
      }
      if (context?.prevPost) {
        queryClient.setQueryData(["post", postId], context.prevPost);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["posts"] });
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      queryClient.invalidateQueries({ queryKey: ["bookmarks"] });
    },
  });
}

async function fetchBookmarks(): Promise<BookmarkListResponse> {
  const res = await fetch("/api/users/bookmarks");
  const json: APIResponse<BookmarkListResponse> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to fetch bookmarks");
  }
  return json.data;
}

export function useBookmarks(enabled = true) {
  return useQuery({
    queryKey: ["bookmarks"],
    queryFn: fetchBookmarks,
    enabled,
  });
}
