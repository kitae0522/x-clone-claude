import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Heart, MessageCircle } from "lucide-react";
import { usePostDetail, useParentChain } from "@/hooks/usePosts";
import { useAuth } from "@/hooks/useAuthContext";
import { useProfile } from "@/hooks/useProfile";
import { useFollow, useUnfollow } from "@/hooks/useFollow";
import { useLike } from "@/hooks/useLike";
import ProfileHoverCard from "@/components/ProfileHoverCard";
import ReplyForm from "@/components/ReplyForm";
import ReplyCard from "@/components/ReplyCard";
import ParentPostCard from "@/components/ParentPostCard";
import { cn } from "@/lib/utils";

export default function PostDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const postId = id ?? "";
  const { data: post, isLoading, error } = usePostDetail(postId);
  const { user: currentUser } = useAuth();
  const [isHoveringFollow, setIsHoveringFollow] = useState(false);

  const authorUsername = post?.author.username ?? "";
  const isOwner = currentUser?.username === authorUsername;
  const { data: authorProfile } = useProfile(
    authorUsername,
    !!post && !isOwner,
  );
  const follow = useFollow(authorUsername);
  const unfollow = useUnfollow(authorUsername);
  const like = useLike(postId, post?.isLiked ?? false);
  const { data: parentChain } = useParentChain(post?.parentId ?? null);

  if (isLoading)
    return (
      <p className="px-4 py-8 text-center text-muted-foreground">Loading...</p>
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
        Post not found.
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
    <div className="mx-auto max-w-[600px]">
      <header className="sticky top-0 z-10 flex items-center gap-4 border-b border-border bg-background/65 px-4 py-3 backdrop-blur-xl">
        <button
          className="cursor-pointer rounded-full border-none bg-none p-1 px-2 text-xl text-foreground transition-colors hover:bg-foreground/10"
          onClick={() => navigate(-1)}
        >
          &larr;
        </button>
        <h1 className="text-xl font-bold">Post</h1>
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
          {post.author.profileImageUrl ? (
            <img
              src={post.author.profileImageUrl}
              alt=""
              className="h-12 w-12 rounded-full object-cover"
            />
          ) : (
            <div className="h-12 w-12 rounded-full bg-border" />
          )}
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
            <button
              onClick={handleFollowClick}
              onMouseEnter={() => setIsHoveringFollow(true)}
              onMouseLeave={() => setIsHoveringFollow(false)}
              className={cn(
                "min-w-[90px] cursor-pointer rounded-full px-3 py-1.5 text-sm font-bold transition-all disabled:cursor-not-allowed disabled:opacity-50",
                authorProfile.isFollowing
                  ? isHoveringFollow
                    ? "border border-destructive/50 bg-transparent text-destructive hover:bg-destructive/10"
                    : "border border-muted-foreground/50 bg-transparent text-foreground"
                  : "border-none bg-foreground text-background hover:bg-foreground/90",
              )}
              disabled={follow.isPending || unfollow.isPending}
            >
              {authorProfile.isFollowing
                ? isHoveringFollow
                  ? "언팔로우"
                  : "팔로잉"
                : "팔로우"}
            </button>
          )}
        </div>
        <p className="mb-4 text-[17px] leading-relaxed text-foreground">
          {post.content}
        </p>
        <div className="flex items-center gap-4 border-t border-border pt-4">
          <span className="text-sm text-muted-foreground">
            {new Date(post.createdAt).toLocaleString()}
          </span>
          <button
            onClick={handleLikeClick}
            disabled={like.isPending}
            className="group flex cursor-pointer items-center gap-1.5 border-none bg-transparent p-0 disabled:opacity-50"
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
            <span
              className={cn(
                "text-sm transition-colors group-hover:text-red-500",
                post.isLiked ? "text-red-500" : "text-muted-foreground",
              )}
            >
              {post.likeCount}
            </span>
          </button>
          <div className="flex items-center gap-1.5 text-muted-foreground">
            <MessageCircle size={20} />
            <span className="text-sm">{post.replyCount}</span>
          </div>
        </div>
      </article>

      <ReplyForm postId={postId} />

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
            No replies yet.
          </p>
        )}
      </section>
    </div>
  );
}
