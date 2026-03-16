import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Bookmark,
  Eye,
  Heart,
  MessageCircle,
  MoreHorizontal,
  Pencil,
  Share,
  Trash2,
} from "lucide-react";
import type { PostDetail } from "@/types/api";
import { useAuth } from "@/hooks/useAuthContext";
import { useLike } from "@/hooks/useLike";
import { useBookmark } from "@/hooks/useBookmark";
import { useDeletePost } from "@/hooks/usePosts";
import ProfileHoverCard from "@/components/ProfileHoverCard";
import UserAvatar from "@/components/UserAvatar";
import ShareModal from "@/components/ShareModal";
import MarkdownRenderer from "@/components/MarkdownRenderer";
import MediaGrid from "@/components/MediaGrid";
import PollDisplay from "@/components/PollDisplay";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { formatCompactNumber } from "@/lib/formatTime";
import { cn } from "@/lib/utils";

interface ReplyCardProps {
  reply: PostDetail;
  parentPostId?: string;
  opAuthorId?: string;
  hasNextSibling?: boolean;
}

export default function ReplyCard({
  reply,
  parentPostId,
  opAuthorId,
  hasNextSibling = false,
}: ReplyCardProps) {
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const like = useLike(reply.id, reply.isLiked, parentPostId);
  const bookmark = useBookmark(reply.id, reply.isBookmarked);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);

  const isOwner = currentUser?.username === reply.author.username;
  const isParentAuthor = opAuthorId != null && currentUser?.id === opAuthorId;
  const canDelete = isOwner || isParentAuthor;
  const isEdited = reply.updatedAt !== reply.createdAt;
  const deletePost = useDeletePost(reply.id);

  const authorThread = reply.topReplies ?? [];
  const hasContinuation = authorThread.length > 0;
  const showLine = hasContinuation || hasNextSibling;
  const isOP = opAuthorId != null && reply.authorId === opAuthorId;

  function handleLikeClick(e: React.MouseEvent) {
    e.stopPropagation();
    if (!currentUser) return;
    like.mutate();
  }

  function handleReplyClick(e: React.MouseEvent) {
    e.stopPropagation();
    if (!currentUser) return;
    navigate(`/compose?replyTo=${reply.id}`);
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
          <div
            className={cn(
              "shrink-0",
              !reply.author.isDeleted && "cursor-pointer",
            )}
            onClick={(e) => {
              e.stopPropagation();
              if (!reply.author.isDeleted)
                navigate(`/${reply.author.username}`);
            }}
          >
            <UserAvatar
              profileImageUrl={reply.author.profileImageUrl}
              displayName={reply.author.displayName || reply.author.username}
              size="md"
            />
          </div>
          {showLine && <div className="mt-1 w-0.5 flex-1 bg-border" />}
        </div>
        <div className="flex-1">
          <div className="mb-1 flex items-center gap-1.5">
            {reply.author.isDeleted ? (
              <span className="text-[14px] font-bold text-muted-foreground">
                {reply.author.displayName || reply.author.username}
              </span>
            ) : (
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
            )}
            {isOP && (
              <span className="rounded-sm bg-primary/15 px-1.5 py-0.5 text-[11px] font-semibold text-primary">
                OP
              </span>
            )}
            <span
              className={cn(
                "text-[13px] text-muted-foreground",
                !reply.author.isDeleted && "cursor-pointer hover:underline",
              )}
              onClick={(e) => {
                e.stopPropagation();
                if (!reply.author.isDeleted)
                  navigate(`/${reply.author.username}`);
              }}
            >
              @{reply.author.username}
            </span>
            <span className="text-[13px] text-muted-foreground">
              · {new Date(reply.createdAt).toLocaleString()}
              {isEdited && (
                <span className="ml-1 text-xs text-muted-foreground/70">
                  (edited)
                </span>
              )}
            </span>
            {currentUser && (
              <div className="ml-auto" onClick={(e) => e.stopPropagation()}>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button className="rounded-full border-none bg-transparent p-1 text-muted-foreground transition-colors hover:bg-primary/10 hover:text-primary cursor-pointer">
                      <MoreHorizontal size={14} />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    {isOwner && (
                      <DropdownMenuItem
                        className="hover:!bg-primary/10 hover:!text-primary data-[highlighted]:!bg-primary data-[highlighted]:!text-white"
                        onClick={() => navigate(`/compose/edit/${reply.id}`)}
                      >
                        <Pencil size={14} className="mr-2" />
                        수정
                      </DropdownMenuItem>
                    )}
                    {canDelete && (
                      <DropdownMenuItem
                        className="text-destructive focus:text-destructive hover:!bg-destructive/10 hover:!text-destructive focus:!bg-destructive/10 data-[highlighted]:!bg-destructive data-[highlighted]:!text-white"
                        onClick={() => setShowDeleteDialog(true)}
                      >
                        <Trash2 size={14} className="mr-2" />
                        삭제
                      </DropdownMenuItem>
                    )}
                    {canDelete && <DropdownMenuSeparator />}
                    <DropdownMenuItem
                      className="hover:!bg-primary/10 hover:!text-primary data-[highlighted]:!bg-primary data-[highlighted]:!text-white"
                      onClick={() => bookmark.mutate()}
                    >
                      <Bookmark
                        size={14}
                        className={cn(
                          "mr-2",
                          reply.isBookmarked && "fill-current",
                        )}
                      />
                      {reply.isBookmarked ? "북마크 제거" : "북마크 추가"}
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      className="hover:!bg-primary/10 hover:!text-primary data-[highlighted]:!bg-primary data-[highlighted]:!text-white"
                      onSelect={(e) => {
                        e.preventDefault();
                        setTimeout(() => setShowShareModal(true), 0);
                      }}
                    >
                      <Share size={14} className="mr-2" />
                      공유하기
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            )}
          </div>
          <div className="mb-2 text-[14px] leading-normal">
            <MarkdownRenderer content={reply.content} />
          </div>
          {reply.media && reply.media.length > 0 && (
            <MediaGrid media={reply.media} />
          )}
          {reply.poll && <PollDisplay poll={reply.poll} postId={reply.id} />}
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
              onClick={handleReplyClick}
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
            <div className="flex items-center gap-1">
              <Eye size={14} className="text-muted-foreground" />
              <span className="text-[12px] text-muted-foreground">
                {reply.viewCount ? formatCompactNumber(reply.viewCount) : ""}
              </span>
            </div>
          </div>
        </div>
      </div>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent onClick={(e) => e.stopPropagation()}>
          <AlertDialogHeader>
            <AlertDialogTitle>답글을 삭제하시겠습니까?</AlertDialogTitle>
            <AlertDialogDescription>
              이 작업은 되돌릴 수 없습니다. 답글이 영구적으로 삭제됩니다.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>취소</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={() => deletePost.mutate()}
            >
              삭제
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <ShareModal
        open={showShareModal}
        onClose={() => setShowShareModal(false)}
        postId={reply.id}
      />

      {authorThread.map((continuation, index) => (
        <ReplyCard
          key={continuation.id}
          reply={continuation}
          parentPostId={reply.id}
          opAuthorId={opAuthorId}
          hasNextSibling={index < authorThread.length - 1}
        />
      ))}
    </>
  );
}
