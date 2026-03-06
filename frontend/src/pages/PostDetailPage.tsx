import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Bookmark,
  Heart,
  MessageCircle,
  Repeat2,
  Share,
  ArrowLeft,
  MapPin,
} from "lucide-react";
import { usePostDetail, useParentChain } from "@/hooks/usePosts";
import { useAuth } from "@/hooks/useAuthContext";
import { useProfile } from "@/hooks/useProfile";
import { useFollow, useUnfollow } from "@/hooks/useFollow";
import { useLike } from "@/hooks/useLike";
import { useBookmark } from "@/hooks/useBookmark";
import { formatRelativeTime, formatCompactNumber } from "@/lib/formatTime";
import ProfileHoverCard from "@/components/ProfileHoverCard";
import UserAvatar from "@/components/UserAvatar";
import { Button } from "@/components/ui/button";

import ReplyCard from "@/components/ReplyCard";
import ParentPostCard from "@/components/ParentPostCard";
import ShareModal from "@/components/ShareModal";
import MarkdownRenderer from "@/components/MarkdownRenderer";
import MediaGrid from "@/components/MediaGrid";
import PollDisplay from "@/components/PollDisplay";
import VisibilityBadge from "@/components/VisibilityBadge";
import { cn } from "@/lib/utils";

export default function PostDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const postId = id ?? "";
  const { data: post, isLoading, error } = usePostDetail(postId);
  const { user: currentUser } = useAuth();
  const [isHoveringFollow, setIsHoveringFollow] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);

  const authorUsername = post?.author.username ?? "";
  const isOwner = currentUser?.username === authorUsername;
  const { data: authorProfile } = useProfile(
    authorUsername,
    !!post && !isOwner,
  );
  const follow = useFollow(authorUsername);
  const unfollow = useUnfollow(authorUsername);
  const like = useLike(postId, post?.isLiked ?? false);
  const bookmark = useBookmark(postId, post?.isBookmarked ?? false);
  const { data: parentChain } = useParentChain(post?.parentId ?? null);

  if (isLoading)
    return (
      <div className="flex justify-center py-8">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  if (error) {
    console.error("PostDetailPage error:", error);
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

  function handleFollowClick() {
    if (authorProfile?.isFollowing) {
      unfollow.mutate();
    } else {
      follow.mutate();
    }
  }

  function handleLikeClick() {
    if (!currentUser) return;
    like.mutate();
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
            className="shrink-0 cursor-pointer"
            onClick={() => navigate(`/${post.author.username}`)}
          >
            <UserAvatar
              profileImageUrl={post.author.profileImageUrl}
              displayName={post.author.displayName || post.author.username}
              size="lg"
            />
          </div>
          <div className="flex flex-1 flex-col">
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
            <span
              className="cursor-pointer text-sm text-muted-foreground hover:underline"
              onClick={(e) => {
                e.stopPropagation();
                navigate(`/${post.author.username}`);
              }}
            >
              @{post.author.username}
            </span>
          </div>
          {!isOwner && currentUser && authorProfile && (
            <Button
              onClick={handleFollowClick}
              onMouseEnter={() => setIsHoveringFollow(true)}
              onMouseLeave={() => setIsHoveringFollow(false)}
              variant={
                authorProfile.isFollowing
                  ? isHoveringFollow
                    ? "follow-danger"
                    : "follow-active"
                  : "follow"
              }
              size="sm"
              className="min-w-[90px] cursor-pointer"
              disabled={follow.isPending || unfollow.isPending}
            >
              {authorProfile.isFollowing
                ? isHoveringFollow
                  ? "언팔로우"
                  : "팔로잉"
                : "팔로우"}
            </Button>
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
          </span>
          <VisibilityBadge visibility={post.visibility} />
        </div>

        {/* Stats */}
        {(post.likeCount > 0 || post.replyCount > 0 || post.viewCount > 0) && (
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
          <button className="group flex cursor-pointer items-center justify-center rounded-full border-none bg-transparent p-2 transition-colors hover:bg-green-500/10">
            <Repeat2
              size={20}
              className="text-muted-foreground transition-colors group-hover:text-green-500"
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
          <button
            onClick={() => {
              if (!currentUser) return;
              bookmark.mutate();
            }}
            disabled={bookmark.isPending}
            className="group flex cursor-pointer items-center justify-center rounded-full border-none bg-transparent p-2 transition-colors hover:bg-primary/10 disabled:opacity-50"
          >
            <Bookmark
              size={20}
              className={cn(
                "transition-colors group-hover:text-primary",
                post.isBookmarked
                  ? "fill-primary text-primary"
                  : "text-muted-foreground",
              )}
            />
          </button>
          <button
            onClick={() => setShowShareModal(true)}
            className="group flex cursor-pointer items-center justify-center rounded-full border-none bg-transparent p-2 transition-colors hover:bg-primary/10"
          >
            <Share
              size={20}
              className="text-muted-foreground transition-colors group-hover:text-primary"
            />
          </button>
        </div>
      </article>

      <ShareModal
        open={showShareModal}
        onClose={() => setShowShareModal(false)}
        postId={postId}
      />

      <section>
        {post.topReplies?.map((reply) => (
          <ReplyCard
            key={reply.id}
            reply={reply}
            parentPostId={postId}
            opUsername={post.author.username}
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
