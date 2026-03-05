import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Heart, MessageCircle } from "lucide-react";
import type { PostDetail } from "@/types/api";
import { useAuth } from "@/hooks/useAuthContext";
import { useLike } from "@/hooks/useLike";
import ProfileHoverCard from "@/components/ProfileHoverCard";
import ReplyForm from "@/components/ReplyForm";
import { cn } from "@/lib/utils";

interface ReplyCardProps {
  reply: PostDetail;
  parentPostId?: string;
  opUsername?: string;
  hasNextSibling?: boolean;
}

export default function ReplyCard({
  reply,
  parentPostId,
  opUsername,
  hasNextSibling = false,
}: ReplyCardProps) {
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const like = useLike(reply.id, reply.isLiked, parentPostId);
  const [showReplyForm, setShowReplyForm] = useState(false);

  const authorThread = reply.topReplies ?? [];
  const hasContinuation = authorThread.length > 0;
  const showLine = hasContinuation || hasNextSibling;
  const isOP = opUsername != null && reply.author.username === opUsername;

  function handleLikeClick(e: React.MouseEvent) {
    e.stopPropagation();
    if (!currentUser) return;
    like.mutate();
  }

  function handleReplyToggle(e: React.MouseEvent) {
    e.stopPropagation();
    if (!currentUser) return;
    setShowReplyForm((prev) => !prev);
  }

  function handleCardClick() {
    navigate(`/post/${reply.id}`);
  }

  return (
    <>
      <div
        onClick={handleCardClick}
        className={cn(
          "flex cursor-pointer gap-3 p-4 transition-colors hover:bg-muted/30",
          !showLine && "border-b border-border",
        )}
      >
        <div className="flex flex-col items-center">
          {reply.author.profileImageUrl ? (
            <img
              src={reply.author.profileImageUrl}
              alt=""
              className="h-10 w-10 rounded-full object-cover"
            />
          ) : (
            <div className="h-10 w-10 rounded-full bg-border" />
          )}
          {showLine && (
            <div className="mt-1 w-0.5 flex-1 bg-border" />
          )}
        </div>
        <div className="flex-1">
          <div className="mb-1 flex items-center gap-1.5">
            <ProfileHoverCard
              handle={reply.author.username}
              currentUsername={currentUser?.username}
            >
              <span
                className="cursor-pointer text-[14px] font-bold text-foreground hover:underline"
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(`/${reply.author.username}`);
                }}
              >
                {reply.author.displayName || reply.author.username}
              </span>
            </ProfileHoverCard>
            {isOP && (
              <span className="rounded-sm bg-primary/15 px-1.5 py-0.5 text-[11px] font-semibold text-primary">
                OP
              </span>
            )}
            <span
              className="cursor-pointer text-[13px] text-muted-foreground hover:underline"
              onClick={(e) => {
                e.stopPropagation();
                navigate(`/${reply.author.username}`);
              }}
            >
              @{reply.author.username}
            </span>
            <span className="text-[13px] text-muted-foreground">
              · {new Date(reply.createdAt).toLocaleString()}
            </span>
          </div>
          <p className="mb-2 text-[14px] leading-normal text-foreground">
            {reply.content}
          </p>
          <div className="flex items-center gap-3">
            <button
              onClick={handleLikeClick}
              disabled={like.isPending}
              className="group flex cursor-pointer items-center gap-1 border-none bg-transparent p-0 disabled:opacity-50"
            >
              <Heart
                size={14}
                className={cn(
                  "transition-colors group-hover:text-red-500",
                  reply.isLiked
                    ? "fill-red-500 text-red-500"
                    : "text-muted-foreground",
                )}
              />
              <span
                className={cn(
                  "text-[12px] transition-colors group-hover:text-red-500",
                  reply.isLiked ? "text-red-500" : "text-muted-foreground",
                )}
              >
                {reply.likeCount}
              </span>
            </button>
            <button
              onClick={handleReplyToggle}
              className="group flex cursor-pointer items-center gap-1 border-none bg-transparent p-0"
            >
              <MessageCircle
                size={14}
                className="text-muted-foreground transition-colors group-hover:text-primary"
              />
              <span className="text-[12px] text-muted-foreground transition-colors group-hover:text-primary">
                {reply.replyCount}
              </span>
            </button>
          </div>
        </div>
      </div>

      {showReplyForm && (
        <ReplyForm postId={reply.id} parentPostId={parentPostId} />
      )}

      {authorThread.map((continuation, index) => (
        <ReplyCard
          key={continuation.id}
          reply={continuation}
          parentPostId={reply.id}
          opUsername={opUsername}
          hasNextSibling={index < authorThread.length - 1}
        />
      ))}
    </>
  );
}
