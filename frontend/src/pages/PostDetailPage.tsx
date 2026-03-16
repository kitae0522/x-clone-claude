import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Bookmark,
  Heart,
  MessageCircle,
  MoreHorizontal,
  Pencil,
  Repeat2,
  Share,
  Trash2,
  ArrowLeft,
  MapPin,
} from "lucide-react";
import { usePostDetail, useParentChain, useDeletePost } from "@/hooks/usePosts";
import { useAuth } from "@/hooks/useAuthContext";
import { useLike } from "@/hooks/useLike";
import { useRepost } from "@/hooks/useRepost";
import { useBookmark } from "@/hooks/useBookmark";
import { formatRelativeTime, formatCompactNumber } from "@/lib/formatTime";
import ProfileHoverCard from "@/components/ProfileHoverCard";
import UserAvatar from "@/components/UserAvatar";

import ReplyCard from "@/components/ReplyCard";
import ParentPostCard from "@/components/ParentPostCard";
import ShareModal from "@/components/ShareModal";
import MarkdownRenderer from "@/components/MarkdownRenderer";
import MediaGrid from "@/components/MediaGrid";
import PollDisplay from "@/components/PollDisplay";
import VisibilityBadge from "@/components/VisibilityBadge";
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

export default function PostDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const postId = id ?? "";
  const { data: post, isLoading, error } = usePostDetail(postId);
  const { user: currentUser } = useAuth();
  const [showShareModal, setShowShareModal] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  const authorUsername = post?.author.username ?? "";
  const isOwner = currentUser?.username === authorUsername;
  const like = useLike(postId, post?.isLiked ?? false);
  const repost = useRepost(postId, post?.isReposted ?? false);
  const bookmark = useBookmark(postId, post?.isBookmarked ?? false);
  const deletePost = useDeletePost(postId);
  const { data: parentChain } = useParentChain(post?.parentId ?? null);
  const isEdited = post ? post.updatedAt !== post.createdAt : false;

  if (isLoading)
    return (
      <div className="flex justify-center py-8">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  if (error) {
    const isDeleted = error.message.includes("410");
    if (isDeleted) {
      return (
        <div className="flex flex-col items-center justify-center px-4 py-16 text-center">
          <Trash2 className="mb-4 h-12 w-12 text-muted-foreground" />
          <h2 className="mb-2 text-lg font-bold">이 게시글은 삭제되었습니다</h2>
          <p className="mb-6 text-sm text-muted-foreground">
            작성자가 이 게시글을 삭제했습니다.
          </p>
          <button
            onClick={() => navigate("/")}
            className="cursor-pointer rounded-full bg-primary px-6 py-2 text-sm font-bold text-primary-foreground transition-colors hover:bg-primary/90 border-none"
          >
            홈으로 돌아가기
          </button>
        </div>
      );
    }
    return (
      <p className="px-4 py-8 text-center text-muted-foreground">
        게시글을 불러오는 중 오류가 발생했습니다.
      </p>
    );
  }
  if (!post)
    return (
      <p className="px-4 py-8 text-center text-muted-foreground">
        게시글을 찾을 수 없습니다.
      </p>
    );

  function handleLikeClick() {
    if (!currentUser) return;
    like.mutate();
  }

  function handleRepostClick() {
    if (!currentUser || isOwner) return;
    repost.mutate();
  }

  return (
    <>
      <header className="sticky top-0 z-10 flex items-center gap-4 border-b border-border bg-background/65 px-4 py-3 backdrop-blur-xl">
        <button
          className="cursor-pointer rounded-full border-none bg-transparent p-2 text-foreground transition-colors hover:bg-foreground/10"
          onClick={() => navigate(-1)}
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h1 className="text-xl font-bold">게시물</h1>
      </header>

      {parentChain && parentChain.length > 0 && (
        <div className="border-b border-border">
          {parentChain.map((parent) => (
            <ParentPostCard key={parent.id} post={parent} />
          ))}
        </div>
      )}

      <article className="p-4">
        <div className="mb-4 flex items-center gap-3">
          <div
            className={cn(
              "shrink-0",
              !post.author.isDeleted && "cursor-pointer",
            )}
            onClick={() => {
              if (!post.author.isDeleted) navigate(`/${post.author.username}`);
            }}
          >
            <UserAvatar
              profileImageUrl={post.author.profileImageUrl}
              displayName={post.author.displayName || post.author.username}
              size="lg"
            />
          </div>
          <div className="flex flex-1 flex-col">
            {post.author.isDeleted ? (
              <span className="text-[15px] font-bold text-muted-foreground">
                {post.author.displayName || post.author.username}
              </span>
            ) : (
              <ProfileHoverCard
                handle={post.author.username}
                currentUsername={currentUser?.username}
              >
                <span
                  className="cursor-pointer text-[15px] font-bold text-foreground hover:underline"
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
                "text-sm text-muted-foreground",
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
          </div>
          {currentUser && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="ml-auto rounded-full border-none bg-transparent p-2 text-muted-foreground transition-colors hover:bg-primary/10 hover:text-primary cursor-pointer">
                  <MoreHorizontal size={18} />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {isOwner && (
                  <>
                    <DropdownMenuItem
                      className="hover:!bg-primary/10 hover:!text-primary data-[highlighted]:!bg-primary data-[highlighted]:!text-white"
                      onClick={() => navigate(`/compose/edit/${postId}`)}
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
                    className={cn("mr-2", post.isBookmarked && "fill-current")}
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

        {/* Location */}
        {post.location && (
          <div className="mb-2 flex items-center gap-1.5 text-[13px] text-muted-foreground">
            <MapPin size={14} />
            <span>{post.location.name}</span>
          </div>
        )}

        {post.content && (
          <div className="mb-4 text-[17px] leading-relaxed">
            <MarkdownRenderer content={post.content} />
          </div>
        )}

        {/* Media */}
        {post.media && post.media.length > 0 && (
          <div className="mb-4">
            <MediaGrid media={post.media} />
          </div>
        )}

        {/* Poll */}
        {post.poll && (
          <div className="mb-4">
            <PollDisplay poll={post.poll} postId={postId} />
          </div>
        )}

        <div className="flex items-center gap-1 text-[15px] text-muted-foreground">
          <span>
            {formatRelativeTime(post.createdAt)} ·{" "}
            {new Date(post.createdAt).toLocaleString("ko-KR")}
            {isEdited && (
              <span className="ml-1 text-xs text-muted-foreground/70">
                (edited)
              </span>
            )}
          </span>
          <VisibilityBadge visibility={post.visibility} />
        </div>

        {/* Stats */}
        {(post.likeCount > 0 ||
          post.replyCount > 0 ||
          post.viewCount > 0 ||
          post.repostCount > 0) && (
          <div className="mt-3 flex gap-4 border-t border-border pt-3 text-sm">
            {post.viewCount > 0 && (
              <span className="text-muted-foreground">
                <strong className="text-foreground">
                  {formatCompactNumber(post.viewCount)}
                </strong>{" "}
                조회
              </span>
            )}
            {post.replyCount > 0 && (
              <span className="text-muted-foreground">
                <strong className="text-foreground">{post.replyCount}</strong>{" "}
                답글
              </span>
            )}
            {post.repostCount > 0 && (
              <span className="text-muted-foreground">
                <strong className="text-foreground">{post.repostCount}</strong>{" "}
                재게시
              </span>
            )}
            {post.likeCount > 0 && (
              <span className="text-muted-foreground">
                <strong className="text-foreground">{post.likeCount}</strong>{" "}
                좋아요
              </span>
            )}
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex items-center justify-around border-t border-border pt-1 mt-3">
          <button
            onClick={() => navigate(`/compose?replyTo=${postId}`)}
            className="group flex cursor-pointer items-center justify-center rounded-full border-none bg-transparent p-2 transition-colors hover:bg-primary/10"
          >
            <MessageCircle
              size={20}
              className="text-muted-foreground transition-colors group-hover:text-primary"
            />
          </button>
          <button
            onClick={handleRepostClick}
            disabled={isOwner}
            className={cn(
              "group flex cursor-pointer items-center justify-center rounded-full border-none bg-transparent p-2 transition-colors hover:bg-green-500/10",
              isOwner && "cursor-not-allowed opacity-50",
            )}
          >
            <Repeat2
              size={20}
              className={cn(
                "transition-colors group-hover:text-green-500",
                post.isReposted ? "text-green-500" : "text-muted-foreground",
              )}
            />
          </button>
          <button
            onClick={handleLikeClick}
            disabled={like.isPending}
            className="group flex cursor-pointer items-center justify-center rounded-full border-none bg-transparent p-2 transition-colors hover:bg-red-500/10 disabled:opacity-50"
          >
            <Heart
              size={20}
              className={cn(
                "transition-colors group-hover:text-red-500",
                post.isLiked
                  ? "fill-red-500 text-red-500"
                  : "text-muted-foreground",
              )}
            />
          </button>
        </div>
      </article>

      <ShareModal
        open={showShareModal}
        onClose={() => setShowShareModal(false)}
        postId={postId}
      />

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {post.parentId
                ? "답글을 삭제하시겠습니까?"
                : "게시글을 삭제하시겠습니까?"}
            </AlertDialogTitle>
            <AlertDialogDescription>
              이 작업은 되돌릴 수 없습니다. {post.parentId ? "답글" : "게시글"}
              이 영구적으로 삭제됩니다.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>취소</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={() => {
                deletePost.mutate(undefined, {
                  onSuccess: () => navigate("/"),
                });
              }}
            >
              삭제
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <section>
        {post.topReplies?.map((reply) => (
          <ReplyCard
            key={reply.id}
            reply={reply}
            parentPostId={postId}
            opAuthorId={post.authorId}
          />
        ))}
        {(!post.topReplies || post.topReplies.length === 0) && (
          <p className="px-4 py-6 text-center text-sm text-muted-foreground">
            아직 답글이 없습니다.
          </p>
        )}
      </section>
    </>
  );
}
