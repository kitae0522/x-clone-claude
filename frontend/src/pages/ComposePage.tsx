import { useNavigate, useSearchParams } from "react-router-dom";
import { useCreatePost, usePostDetail } from "@/hooks/usePosts";
import { useCreateReply } from "@/hooks/useReplies";
import { toast } from "sonner";
import UserAvatar from "@/components/UserAvatar";
import MarkdownRenderer from "@/components/MarkdownRenderer";
import MediaGrid from "@/components/MediaGrid";
import PollDisplay from "@/components/PollDisplay";
import ComposeLayout from "@/components/ComposeLayout";
import type { ComposeSubmitData } from "@/components/ComposeLayout";

export default function ComposePage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const replyToId = searchParams.get("replyTo");

  const { mutate: createPost, isPending: isPostPending } = useCreatePost();
  const { mutate: createReply, isPending: isReplyPending } = useCreateReply(
    replyToId ?? "",
  );
  const { data: parentPost, isLoading: isParentLoading } = usePostDetail(
    replyToId ?? "",
  );

  function handleSubmit(data: ComposeSubmitData) {
    if (replyToId) {
      createReply(
        {
          content: data.content,
          mediaIds:
            data.mediaIds.length > 0 ? data.mediaIds : undefined,
          location: data.location ?? undefined,
          poll: data.poll ?? undefined,
        },
        {
          onSuccess: () => {
            toast.success("답글이 작성되었습니다.");
            navigate(`/post/${replyToId}`);
          },
          onError: (err) => {
            toast.error("답글 작성에 실패했습니다.", {
              description: err.message,
            });
          },
        },
      );
    } else {
      createPost(
        {
          content: data.content,
          visibility: data.visibility,
          mediaIds:
            data.mediaIds.length > 0 ? data.mediaIds : undefined,
          location: data.location ?? undefined,
          poll: data.poll ?? undefined,
        },
        {
          onSuccess: () => {
            toast.success("게시글이 작성되었습니다.");
            navigate("/");
          },
          onError: (err) => {
            toast.error("게시글 작성에 실패했습니다.", {
              description: err.message,
            });
          },
        },
      );
    }
  }

  return (
    <ComposeLayout
      onSubmit={handleSubmit}
      isPending={replyToId ? isReplyPending : isPostPending}
      submitLabel={replyToId ? "답글 작성" : "게시하기"}
      pendingLabel="게시 중..."
      placeholder={
        replyToId ? "답글을 입력하세요..." : "무슨 일이 일어나고 있나요?"
      }
      showVisibility={!replyToId}
    >
      {/* Reply context */}
      {replyToId && parentPost && (
        <div className="border-border px-4 pt-3">
          <div className="flex gap-3">
            <div className="flex flex-col items-center">
              <UserAvatar
                profileImageUrl={parentPost.author.profileImageUrl}
                displayName={
                  parentPost.author.displayName || parentPost.author.username
                }
                size="md"
              />
              <div className="mt-1 w-0.5 flex-1 bg-border" />
            </div>
            <div className="min-w-0 flex-1 pb-3">
              <div className="flex items-center gap-1">
                <span className="text-[15px] font-bold">
                  {parentPost.author.displayName || parentPost.author.username}
                </span>
                <span className="text-[15px] text-muted-foreground">
                  @{parentPost.author.username}
                </span>
              </div>
              <div className="mt-1 text-[15px] leading-normal">
                <MarkdownRenderer content={parentPost.content} />
              </div>
              {parentPost.media && parentPost.media.length > 0 && (
                <MediaGrid media={parentPost.media} />
              )}
              {parentPost.poll && (
                <PollDisplay poll={parentPost.poll} postId={parentPost.id} />
              )}
              <div className="mt-2 text-[13px] text-muted-foreground">
                <span className="text-primary">
                  @{parentPost.author.username}
                </span>
                님에게 답글 남기는 중
              </div>
            </div>
          </div>
        </div>
      )}

      {replyToId && isParentLoading && (
        <div className="flex justify-center border-b border-border py-6">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      )}
    </ComposeLayout>
  );
}
