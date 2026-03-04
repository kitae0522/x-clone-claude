import { useNavigate } from 'react-router-dom'
import { Heart } from 'lucide-react'
import type { PostDetail } from '@/types/api'
import { useAuth } from '@/hooks/useAuthContext'
import { useLike } from '@/hooks/useLike'
import ProfileHoverCard from '@/components/ProfileHoverCard'
import { cn } from '@/lib/utils'

interface ReplyCardProps {
  reply: PostDetail
}

export default function ReplyCard({ reply }: ReplyCardProps) {
  const navigate = useNavigate()
  const { user: currentUser } = useAuth()
  const like = useLike(reply.id, reply.isLiked)

  function handleLikeClick(e: React.MouseEvent) {
    e.stopPropagation()
    if (!currentUser) return
    like.mutate()
  }

  return (
    <div className="flex gap-3 border-b border-border p-4 pl-6">
      <div className="flex flex-col items-center">
        {reply.author.profileImageUrl ? (
          <img
            src={reply.author.profileImageUrl}
            alt=""
            className="h-8 w-8 rounded-full object-cover"
          />
        ) : (
          <div className="h-8 w-8 rounded-full bg-border" />
        )}
        <div className="mt-1 w-0.5 flex-1 bg-border" />
      </div>
      <div className="flex-1">
        <div className="mb-1 flex items-center gap-1.5">
          <ProfileHoverCard
            handle={reply.author.username}
            currentUsername={currentUser?.username}
          >
            <span
              className="cursor-pointer text-[14px] font-bold text-foreground hover:underline"
              onClick={() => navigate(`/${reply.author.username}`)}
            >
              {reply.author.displayName || reply.author.username}
            </span>
          </ProfileHoverCard>
          <span
            className="cursor-pointer text-[13px] text-muted-foreground hover:underline"
            onClick={() => navigate(`/${reply.author.username}`)}
          >
            @{reply.author.username}
          </span>
          <span className="text-[13px] text-muted-foreground">
            · {new Date(reply.createdAt).toLocaleString()}
          </span>
        </div>
        <p className="mb-2 text-[14px] leading-normal text-foreground">{reply.content}</p>
        <button
          onClick={handleLikeClick}
          className="group flex cursor-pointer items-center gap-1 border-none bg-transparent p-0"
        >
          <Heart
            size={14}
            className={cn(
              'transition-colors group-hover:text-red-500',
              reply.isLiked ? 'fill-red-500 text-red-500' : 'text-muted-foreground',
            )}
          />
          <span
            className={cn(
              'text-[12px] transition-colors group-hover:text-red-500',
              reply.isLiked ? 'text-red-500' : 'text-muted-foreground',
            )}
          >
            {reply.likeCount}
          </span>
        </button>
      </div>
    </div>
  )
}
