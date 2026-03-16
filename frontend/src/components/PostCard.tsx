import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Bookmark,
  Eye,
  Heart,
  MessageCircle,
  MoreHorizontal,
  Pencil,
  Repeat2,
  Share,
  Trash2,
  MapPin,
} from "lucide-react";
import VisibilityBadge from "@/components/VisibilityBadge";
import type { PostDetail } from "@/types/api";
import { useAuth } from "@/hooks/useAuthContext";
import { useLike } from "@/hooks/useLike";
import { useRepost } from "@/hooks/useRepost";
import { useBookmark } from "@/hooks/useBookmark";
import { formatRelativeTime, formatCompactNumber } from "@/lib/formatTime";
import ProfileHoverCard from "@/components/ProfileHoverCard";
import ShareModal from "@/components/ShareModal";
import UserAvatar from "@/components/UserAvatar";
import MarkdownRenderer from "@/components/MarkdownRenderer";
import MediaGrid from "@/components/MediaGrid";
import PollDisplay from "@/components/PollDisplay";
import { useDeletePost } from "@/hooks/usePosts";
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
import { cn } from "@/lib/utils";

interface PostCardProps {
  post: PostDetail;
}

function PostCard({ post }: PostCardProps) {
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const [showShareModal, setShowShareModal] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  const isOwner = currentUser?.username === post.author.username;
  const isEdited = post.updatedAt !== post.createdAt;

  const deletePost = useDeletePost(post.id);
  const like = useLike(post.id, post.isLiked);
  const repost = useRepost(post.id, post.isReposted);
  const bookmark = useBookmark(post.id, post.isBookmarked);

  function handleLikeClick(e: React.MouseEvent) {
    e.stopPropagation();
    if (!currentUser) return;
    like.mutate();
  }

  function handleRepostClick(e: React.MouseEvent) {
    e.stopPropagation();
    if (!currentUser || isOwner) return;
    repost.mutate();
  }

  return (
    <article
      className="cursor-pointer border-b border-border px-4 py-3 transition-colors hover:bg-foreground/[0.03]"
      onClick={() => navigate(`/post/${post.id}`)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "Enter") navigate(`/post/${post.id}`);
      }}
    >
      {post.repostedBy && (
        <div className="flex items-center gap-1.5 px-4 pb-1 text-[13px] text-muted-foreground">
          <Repeat2 size={14} className="text-green-500" />
          <span>
            <ProfileHoverCard
              handle={post.repostedBy.username}
              currentUsername={currentUser?.username}
            >
              <span
                className="cursor-pointer hover:underline"
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(`/${post.repostedBy!.username}`);
                }}
              >
                {post.repostedBy.displayName || post.repostedBy.username}
              </span>
            </ProfileHoverCard>
            님이 재게시함
          </span>
        </div>
      )}
      <div className="flex gap-3">
        <div
          className={cn(
            "mt-0.5 shrink-0",
            !post.author.isDeleted && "cursor-pointer",
          )}
          onClick={(e) => {
            e.stopPropagation();
            if (!post.author.isDeleted) navigate(`/${post.author.username}`);
          }}
        >
          <UserAvatar
            profileImageUrl={post.author.profileImageUrl}
            displayName={post.author.displayName || post.author.username}
            size="md"
          />
        </div>
        <div className="min-w-0 flex-1">
          {/* Author Row */}
          <div className="flex items-center justify-between">
            <div className="flex min-w-0 items-center gap-1">
              {post.author.isDeleted ? (
                <span className="truncate text-[15px] font-bold text-muted-foreground">
                  {post.author.displayName || post.author.username}
                </span>
              ) : (
                <ProfileHoverCard
                  handle={post.author.username}
                  currentUsername={currentUser?.username}
                >
                  <span
                    className="cursor-pointer truncate text-[15px] font-bold text-foreground hover:underline"
                    onClick={(e) => {
                      e.stopPropagation();
                      navigate(`/${post.author.username}`);
                    }}
                  >
                    {post.author.displayName || post.author.username}
                  </span>
                </ProfileHoverCard>
              )}
              <span
                className={cn(
                  "text-[15px] text-muted-foreground",
                  !post.author.isDeleted && "cursor-pointer hover:underline",
                )}
                onClick={(e) => {
                  e.stopPropagation();
                  if (!post.author.isDeleted)
                    navigate(`/${post.author.username}`);
                }}
              >
                @{post.author.username}
              </span>
              <span className="text-muted-foreground">·</span>
              <span className="shrink-0 text-[15px] text-muted-foreground">
                {formatRelativeTime(post.createdAt)}
                {isEdited && (
                  <span className="ml-1 text-xs text-muted-foreground/70">
                    (edited)
                  </span>
                )}
              </span>
              <VisibilityBadge visibility={post.visibility} />
            </div>
            {currentUser && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    onClick={(e) => e.stopPropagation()}
                    className="ml-auto rounded-full border-none bg-transparent p-1.5 text-muted-foreground transition-colors hover:bg-primary/10 hover:text-primary cursor-pointer"
                  >
                    <MoreHorizontal size={16} />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  onClick={(e) => e.stopPropagation()}
                >
                  {isOwner && (
                    <>
                      <DropdownMenuItem
                        className="hover:!bg-primary/10 hover:!text-primary data-[highlighted]:!bg-primary data-[highlighted]:!text-white"
                        onClick={() => navigate(`/compose/edit/${post.id}`)}
                      >
                        <Pencil size={14} className="mr-2" />
                        수정
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="text-destructive focus:text-destructive hover:!bg-destructive/10 hover:!text-destructive focus:!bg-destructive/10 data-[highlighted]:!bg-destructive data-[highlighted]:!text-white"
                        onClick={() => setShowDeleteDialog(true)}
                      >
                        <Trash2 size={14} className="mr-2" />
                        삭제
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                    </>
                  )}
                  <DropdownMenuItem
                    className="hover:!bg-primary/10 hover:!text-primary data-[highlighted]:!bg-primary data-[highlighted]:!text-white"
                    onClick={() => bookmark.mutate()}
                  >
                    <Bookmark
                      size={14}
                      className={cn(
                        "mr-2",
                        post.isBookmarked && "fill-current",
                      )}
                    />
                    {post.isBookmarked ? "북마크 제거" : "북마크 추가"}
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
            )}
          </div>

          {/* Replying to context */}
          {post.parent && (
            <div
              className="mt-0.5 flex items-center gap-1 text-[13px] text-muted-foreground"
              onClick={(e) => {
                e.stopPropagation();
                navigate(`/post/${post.parent!.id}`);
              }}
            >
              <span>
                <span className="text-muted-foreground">replying to </span>
                <span
                  className={cn(
                    post.parent.author.isDeleted
                      ? "text-muted-foreground"
                      : "cursor-pointer text-primary hover:underline",
                  )}
                >
                  @{post.parent.author.username}
                </span>
              </span>
              <span className="truncate text-muted-foreground/70">
                — {post.parent.content}
              </span>
            </div>
          )}

          {/* Location */}
          {post.location && (
            <div className="mt-0.5 flex items-center gap-1 text-[13px] text-muted-foreground">
              <MapPin size={12} />
              <span>{post.location.name}</span>
            </div>
          )}

          {/* Content */}
          {post.content && (
            <div className="mt-0.5 text-[15px] leading-normal">
              <MarkdownRenderer content={post.content} />
            </div>
          )}

          {/* Media */}
          {post.media && post.media.length > 0 && (
            <MediaGrid media={post.media} />
          )}

          {/* Poll */}
          {post.poll && <PollDisplay poll={post.poll} postId={post.id} />}

          {/* Action Buttons */}
          <div className="-ml-2 mt-1.5 flex max-w-[425px] items-center justify-between">
            {/* Reply */}
            <button
              onClick={(e) => {
                e.stopPropagation();
                navigate(`/compose?replyTo=${post.id}`);
              }}
              className="group flex cursor-pointer items-center gap-1.5 rounded-full border-none bg-transparent p-2 transition-colors hover:bg-primary/10"
            >
              <MessageCircle
                size={18}
                className="text-muted-foreground transition-colors group-hover:text-primary"
              />
              <span className="text-[13px] text-muted-foreground transition-colors group-hover:text-primary">
                {post.replyCount || ""}
              </span>
            </button>

            {/* Repost */}
            <button
              onClick={handleRepostClick}
              disabled={isOwner}
              className={cn(
                "group flex cursor-pointer items-center gap-1.5 rounded-full border-none bg-transparent p-2 transition-colors hover:bg-green-500/10",
                isOwner && "cursor-not-allowed opacity-50",
              )}
            >
              <Repeat2
                size={18}
                className={cn(
                  "transition-colors group-hover:text-green-500",
                  post.isReposted ? "text-green-500" : "text-muted-foreground",
                )}
              />
              <span
                className={cn(
                  "text-[13px] transition-colors group-hover:text-green-500",
                  post.isReposted ? "text-green-500" : "text-muted-foreground",
                )}
              >
                {post.repostCount || ""}
              </span>
            </button>

            {/* Like */}
            <button
              onClick={handleLikeClick}
              className="group flex cursor-pointer items-center gap-1.5 rounded-full border-none bg-transparent p-2 transition-colors hover:bg-red-500/10"
            >
              <Heart
                size={18}
                className={cn(
                  "transition-colors group-hover:text-red-500",
                  post.isLiked
                    ? "fill-red-500 text-red-500"
                    : "text-muted-foreground",
                )}
              />
              <span
                className={cn(
                  "text-[13px] transition-colors group-hover:text-red-500",
                  post.isLiked ? "text-red-500" : "text-muted-foreground",
                )}
              >
                {post.likeCount || ""}
              </span>
            </button>

            {/* View Count */}
            <div className="group flex items-center gap-1.5 rounded-full p-2">
              <Eye size={18} className="text-muted-foreground" />
              <span className="text-[13px] text-muted-foreground">
                {post.viewCount ? formatCompactNumber(post.viewCount) : ""}
              </span>
            </div>
          </div>

          <ShareModal
            open={showShareModal}
            onClose={() => setShowShareModal(false)}
            postId={post.id}
          />

          <AlertDialog
            open={showDeleteDialog}
            onOpenChange={setShowDeleteDialog}
          >
            <AlertDialogContent onClick={(e) => e.stopPropagation()}>
              <AlertDialogHeader>
                <AlertDialogTitle>
                  {post.parentId
                    ? "답글을 삭제하시겠습니까?"
                    : "게시글을 삭제하시겠습니까?"}
                </AlertDialogTitle>
                <AlertDialogDescription>
                  이 작업은 되돌릴 수 없습니다.{" "}
                  {post.parentId ? "답글" : "게시글"}이 영구적으로 삭제됩니다.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>취소</AlertDialogCancel>
                <AlertDialogAction
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  onClick={() => {
                    deletePost.mutate();
                  }}
                >
                  삭제
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>
    </article>
  );
}

export default PostCard;
