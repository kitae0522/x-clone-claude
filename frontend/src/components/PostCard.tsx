import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Heart, MessageCircle } from 'lucide-react'
import type { PostDetail } from '@/types/api'
import { useAuth } from '@/hooks/useAuthContext'
import { useProfile } from '@/hooks/useProfile'
import { useFollow, useUnfollow } from '@/hooks/useFollow'
import { useLike } from '@/hooks/useLike'
import ProfileHoverCard from '@/components/ProfileHoverCard'
import { cn } from '@/lib/utils'

interface PostCardProps {
  post: PostDetail
}

const visibilityLabel: Record<PostDetail['visibility'], string> = {
  public: 'Public',
  friends: 'Friends',
  private: 'Private',
}

const visibilityClasses: Record<PostDetail['visibility'], string> = {
  public: 'bg-green-900/50 text-green-400',
  friends: 'bg-blue-900/50 text-blue-400',
  private: 'bg-red-900/50 text-red-400',
}

function PostCard({ post }: PostCardProps) {
  const navigate = useNavigate()
  const { user: currentUser } = useAuth()
  const [isHoveringFollow, setIsHoveringFollow] = useState(false)

  const isOwner = currentUser?.username === post.author.username
  const { data: authorProfile } = useProfile(post.author.username, !isOwner)
  const follow = useFollow(post.author.username)
  const unfollow = useUnfollow(post.author.username)
  const like = useLike(post.id, post.isLiked)

  function handleFollowClick(e: React.MouseEvent) {
    e.stopPropagation()
    if (authorProfile?.isFollowing) {
      unfollow.mutate()
    } else {
      follow.mutate()
    }
  }

  function handleLikeClick(e: React.MouseEvent) {
    e.stopPropagation()
    if (!currentUser) return
    like.mutate()
  }

  return (
    <div
      className="cursor-pointer border-b border-border p-4 transition-colors hover:bg-foreground/[0.03]"
      onClick={() => navigate(`/post/${post.id}`)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter') navigate(`/post/${post.id}`)
      }}
    >
      <div className="mb-2 flex items-center justify-between">
        <div className="flex items-center gap-2.5">
          {post.author.profileImageUrl ? (
            <img
              src={post.author.profileImageUrl}
              alt=""
              className="h-10 w-10 rounded-full object-cover"
            />
          ) : (
            <div className="h-10 w-10 rounded-full bg-border" />
          )}
          <div>
            <ProfileHoverCard
              handle={post.author.username}
              currentUsername={currentUser?.username}
            >
              <span
                className="mr-1 cursor-pointer text-[15px] font-bold text-foreground hover:underline"
                onClick={(e) => {
                  e.stopPropagation()
                  navigate(`/${post.author.username}`)
                }}
              >
                {post.author.displayName || post.author.username}
              </span>
            </ProfileHoverCard>
            <span
              className="cursor-pointer text-sm text-muted-foreground hover:underline"
              onClick={(e) => {
                e.stopPropagation()
                navigate(`/${post.author.username}`)
              }}
            >
              @{post.author.username}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {!isOwner && currentUser && authorProfile && (
            <button
              onClick={handleFollowClick}
              onMouseEnter={() => setIsHoveringFollow(true)}
              onMouseLeave={() => setIsHoveringFollow(false)}
              className={cn(
                'min-w-[90px] cursor-pointer rounded-full px-3 py-1 text-[13px] font-bold transition-all disabled:cursor-not-allowed disabled:opacity-50',
                authorProfile.isFollowing
                  ? isHoveringFollow
                    ? 'border border-destructive/50 bg-transparent text-destructive hover:bg-destructive/10'
                    : 'border border-muted-foreground/50 bg-transparent text-foreground'
                  : 'border-none bg-foreground text-background hover:bg-foreground/90',
              )}
              disabled={follow.isPending || unfollow.isPending}
            >
              {authorProfile.isFollowing
                ? isHoveringFollow
                  ? '언팔로우'
                  : '팔로잉'
                : '팔로우'}
            </button>
          )}
          <span className={cn('rounded-full px-2 py-0.5 text-xs font-medium', visibilityClasses[post.visibility])}>
            {visibilityLabel[post.visibility]}
          </span>
        </div>
      </div>
      <p className="mb-2 text-[15px] leading-normal text-foreground">{post.content}</p>
      <div className="flex items-center gap-4">
        <span className="text-[13px] text-muted-foreground">
          {new Date(post.createdAt).toLocaleString()}
        </span>
        <div className="flex items-center gap-1 text-muted-foreground">
          <MessageCircle size={16} />
          <span className="text-[13px]">{post.replyCount}</span>
        </div>
        <button
          onClick={handleLikeClick}
          className="group flex cursor-pointer items-center gap-1 border-none bg-transparent p-0"
        >
          <Heart
            size={16}
            className={cn(
              'transition-colors group-hover:text-red-500',
              post.isLiked ? 'fill-red-500 text-red-500' : 'text-muted-foreground',
            )}
          />
          <span
            className={cn(
              'text-[13px] transition-colors group-hover:text-red-500',
              post.isLiked ? 'text-red-500' : 'text-muted-foreground',
            )}
          >
            {post.likeCount}
          </span>
        </button>
      </div>
    </div>
  )
}

export default PostCard
