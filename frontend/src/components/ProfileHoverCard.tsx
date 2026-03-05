import { useState, useRef, useCallback, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { useProfile } from '@/hooks/useProfile'
import { useFollow, useUnfollow } from '@/hooks/useFollow'
import UserAvatar from '@/components/UserAvatar'
import { Button } from '@/components/ui/button'

interface Props {
  handle: string
  currentUsername?: string
  children: ReactNode
}

export default function ProfileHoverCard({ handle, currentUsername, children }: Props) {
  const [isOpen, setIsOpen] = useState(false)
  const [isHoveringFollow, setIsHoveringFollow] = useState(false)
  const openTimeout = useRef<ReturnType<typeof setTimeout>>(null)
  const closeTimeout = useRef<ReturnType<typeof setTimeout>>(null)
  const navigate = useNavigate()

  const { data: profile } = useProfile(handle, isOpen)
  const follow = useFollow(handle)
  const unfollow = useUnfollow(handle)

  const isOwner = currentUsername === handle

  const handleMouseEnter = useCallback(() => {
    if (closeTimeout.current) clearTimeout(closeTimeout.current)
    openTimeout.current = setTimeout(() => setIsOpen(true), 300)
  }, [])

  const handleMouseLeave = useCallback(() => {
    if (openTimeout.current) clearTimeout(openTimeout.current)
    closeTimeout.current = setTimeout(() => setIsOpen(false), 200)
  }, [])

  function handleFollowClick(e: React.MouseEvent) {
    e.stopPropagation()
    if (profile?.isFollowing) {
      unfollow.mutate()
    } else {
      follow.mutate()
    }
  }

  return (
    <div
      className="relative inline-block"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {children}
      {isOpen && profile && (
        <div className="absolute left-0 top-full z-50 mt-2 w-[300px] rounded-2xl border border-border bg-background p-4 shadow-lg">
          <div className="flex items-start justify-between">
            <div
              className="cursor-pointer"
              onClick={(e) => {
                e.stopPropagation()
                navigate(`/${profile.username}`)
              }}
            >
              <UserAvatar
                profileImageUrl={profile.profileImageUrl}
                displayName={profile.displayName}
                size="xl"
              />
            </div>
            {!isOwner && (
              <Button
                onClick={handleFollowClick}
                onMouseEnter={() => setIsHoveringFollow(true)}
                onMouseLeave={() => setIsHoveringFollow(false)}
                variant={
                  profile.isFollowing
                    ? isHoveringFollow
                      ? 'follow-danger'
                      : 'follow-active'
                    : 'follow'
                }
                size="sm"
                className="min-w-[90px] cursor-pointer"
                disabled={follow.isPending || unfollow.isPending}
              >
                {profile.isFollowing
                  ? isHoveringFollow
                    ? '언팔로우'
                    : '팔로잉'
                  : '팔로우'}
              </Button>
            )}
          </div>
          <div className="mt-2">
            <div
              className="cursor-pointer text-[15px] font-bold hover:underline"
              onClick={(e) => {
                e.stopPropagation()
                navigate(`/${profile.username}`)
              }}
            >
              {profile.displayName}
            </div>
            <div className="text-[15px] text-muted-foreground">@{profile.username}</div>
          </div>
          {profile.bio && (
            <p className="mt-2 whitespace-pre-wrap text-[15px] leading-relaxed">{profile.bio}</p>
          )}
          <div className="mt-2 flex gap-4">
            <span className="text-sm text-muted-foreground">
              <strong className="text-foreground">{profile.followingCount}</strong> 팔로잉
            </span>
            <span className="text-sm text-muted-foreground">
              <strong className="text-foreground">{profile.followersCount}</strong> 팔로워
            </span>
          </div>
        </div>
      )}
    </div>
  )
}
