import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Bookmark, Heart, MessageCircle, Repeat2, Share, MapPin } from "lucide-react";
import type { PostDetail } from "@/types/api";
import { useAuth } from "@/hooks/useAuthContext";
import { useProfile } from "@/hooks/useProfile";
import { useFollow, useUnfollow } from "@/hooks/useFollow";
import { useLike } from "@/hooks/useLike";
import { useBookmark } from "@/hooks/useBookmark";
import { formatRelativeTime } from "@/lib/formatTime";
import ProfileHoverCard from "@/components/ProfileHoverCard";
import ShareModal from "@/components/ShareModal";
import UserAvatar from "@/components/UserAvatar";
import MarkdownRenderer from "@/components/MarkdownRenderer";
import MediaGrid from "@/components/MediaGrid";
import PollDisplay from "@/components/PollDisplay";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface PostCardProps {
  post: PostDetail;
}

function PostCard({ post }: PostCardProps) {
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const [isHoveringFollow, setIsHoveringFollow] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);

  const isOwner = currentUser?.username === post.author.username;
  const { data: authorProfile } = useProfile(post.author.username, !isOwner);
  const follow = useFollow(post.author.username);
  const unfollow = useUnfollow(post.author.username);
  const like = useLike(post.id, post.isLiked);
  const bookmark = useBookmark(post.id, post.isBookmarked);

  function handleFollowClick(e: React.MouseEvent) {
    e.stopPropagation();
    if (authorProfile?.isFollowing) {
      unfollow.mutate();
    } else {
      follow.mutate();
    }
  }

  function handleLikeClick(e: React.MouseEvent) {
    e.stopPropagation();
    if (!currentUser) return;
    like.mutate();
  }

  function handleBookmarkClick(e: React.MouseEvent) {
    e.stopPropagation();
    if (!currentUser) return;
    bookmark.mutate();
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
      <div className="flex gap-3">
        <div
          className="mt-0.5 shrink-0 cursor-pointer"
          onClick={(e) => {
            e.stopPropagation()
            navigate(`/${post.author.username}`)
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
              <span
                className="cursor-pointer text-[15px] text-muted-foreground hover:underline"
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(`/${post.author.username}`);
                }}
              >
                @{post.author.username}
              </span>
              <span className="text-muted-foreground">·</span>
              <span className="shrink-0 text-[15px] text-muted-foreground">
                {formatRelativeTime(post.createdAt)}
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
                className="ml-2 min-w-[80px] cursor-pointer"
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
                <span className="cursor-pointer text-primary hover:underline">
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
          {post.poll && (
            <PollDisplay
              poll={post.poll}
              postId={post.id}
              isOwnPost={currentUser?.username === post.author.username}
            />
          )}

          {/* Action Buttons */}
          <div className="-ml-2 mt-1.5 flex max-w-[425px] items-center justify-between">
            {/* Reply */}
            <button
              onClick={(e) => {
                e.stopPropagation();
                navigate(`/post/${post.id}`);
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
              onClick={(e) => e.stopPropagation()}
              className="group flex cursor-pointer items-center gap-1.5 rounded-full border-none bg-transparent p-2 transition-colors hover:bg-green-500/10"
            >
              <Repeat2
                size={18}
                className="text-muted-foreground transition-colors group-hover:text-green-500"
              />
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

            {/* Bookmark */}
            <button
              onClick={handleBookmarkClick}
              className="group flex cursor-pointer items-center gap-1.5 rounded-full border-none bg-transparent p-2 transition-colors hover:bg-primary/10"
            >
              <Bookmark
                size={18}
                className={cn(
                  "transition-colors group-hover:text-primary",
                  post.isBookmarked
                    ? "fill-primary text-primary"
                    : "text-muted-foreground",
                )}
              />
            </button>

            {/* Share */}
            <button
              onClick={(e) => {
                e.stopPropagation();
                setShowShareModal(true);
              }}
              className="group flex cursor-pointer items-center gap-1.5 rounded-full border-none bg-transparent p-2 transition-colors hover:bg-primary/10"
            >
              <Share
                size={18}
                className="text-muted-foreground transition-colors group-hover:text-primary"
              />
            </button>
          </div>

          <ShareModal
            open={showShareModal}
            onClose={() => setShowShareModal(false)}
            postId={post.id}
          />
        </div>
      </div>
    </article>
  );
}

export default PostCard;
