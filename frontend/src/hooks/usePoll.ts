import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { APIResponse, PollData } from "@/types/api";

interface VoteRequest {
  postId: string;
  optionIndex: number;
}

async function vote({ postId, optionIndex }: VoteRequest): Promise<PollData> {
  const res = await fetch(`/api/posts/${postId}/vote`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ optionIndex }),
  });
  if (!res.ok) {
    const json = await res.json();
    throw new Error(json.error ?? `Vote failed: ${res.status}`);
  }
  const json: APIResponse<{ poll: PollData }> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Unknown error");
  }
  return json.data.poll;
}

export function useVote(postId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (optionIndex: number) => vote({ postId, optionIndex }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      queryClient.invalidateQueries({ queryKey: ["posts"] });
    },
  });
}

async function unvote(postId: string): Promise<PollData> {
  const res = await fetch(`/api/posts/${postId}/vote`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!res.ok) {
    const json = await res.json();
    throw new Error(json.error ?? `Unvote failed: ${res.status}`);
  }
  const json: APIResponse<{ poll: PollData }> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Unknown error");
  }
  return json.data.poll;
}

export function useUnvote(postId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => unvote(postId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["post", postId] });
      queryClient.invalidateQueries({ queryKey: ["posts"] });
    },
  });
}
