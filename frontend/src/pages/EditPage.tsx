import { useParams, useNavigate } from "react-router-dom";
import { usePostDetail, useUpdatePost } from "@/hooks/usePosts";
import { toast } from "sonner";
import ComposeLayout from "@/components/ComposeLayout";
import type { ComposeSubmitData } from "@/components/ComposeLayout";
import type { UpdatePostRequest } from "@/types/api";

export default function EditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const postId = id ?? "";
  const { data: post, isLoading, error } = usePostDetail(postId);
  const updatePost = useUpdatePost(postId);

  if (isLoading) {
    return (
      <div className="flex justify-center py-8">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  if (error || !post) {
    return (
      <p className="px-4 py-8 text-center text-muted-foreground">
        게시글을 불러오는 중 오류가 발생했습니다.
      </p>
    );
  }

  function handleSubmit(data: ComposeSubmitData) {
    const req: UpdatePostRequest = {
      content: data.content,
    };

    if (!post!.parentId) {
      req.visibility = data.visibility;
    }

    // 미디어가 실제로 변경된 경우에만 mediaIds 전송
    const initialMediaIds = (post!.media ?? []).map((m) => m.id).sort();
    const currentMediaIds = [...data.mediaIds].sort();
    const mediaChanged =
      initialMediaIds.length !== currentMediaIds.length ||
      initialMediaIds.some((id, i) => id !== currentMediaIds[i]);
    if (mediaChanged) {
      req.mediaIds = data.mediaIds;
    }

    if (data.location) {
      req.location = data.location;
      req.clearLocation = false;
    } else if (post!.location) {
      req.clearLocation = true;
    }

    if (data.poll) {
      req.poll = data.poll;
      req.clearPoll = false;
    } else if (post!.poll) {
      req.clearPoll = true;
    }

    updatePost.mutate(req, {
      onSuccess: () => {
        toast.success("수정되었습니다.");
        navigate(`/post/${postId}`);
      },
      onError: (err) => {
        toast.error("수정에 실패했습니다.", { description: err.message });
      },
    });
  }

  return (
    <ComposeLayout
      onSubmit={handleSubmit}
      isPending={updatePost.isPending}
      submitLabel="수정하기"
      pendingLabel="수정 중..."
      showVisibility={!post.parentId}
      initialContent={post.content}
      initialVisibility={post.visibility}
      initialLocation={post.location}
      initialMediaItems={post.media ?? []}
      initialPollOptions={
        post.poll ? post.poll.options.map((o) => o.text) : undefined
      }
      initialPollDuration={1440}
    />
  );
}
