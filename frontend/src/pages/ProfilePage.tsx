import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useProfile } from '@/hooks/useProfile'
import { useAuth } from '@/hooks/useAuthContext'
import { useFollow, useUnfollow } from '@/hooks/useFollow'
import EditProfileModal from '@/components/EditProfileModal'
import FollowListModal from '@/components/FollowListModal'
import { cn } from '@/lib/utils'

export default function ProfilePage() {
  const { handle } = useParams<{ handle: string }>()
  const navigate = useNavigate()
  const { user: currentUser } = useAuth()
  const { data: profile, isLoading, error } = useProfile(handle ?? '')
  const [showEditModal, setShowEditModal] = useState(false)
  const [followListType, setFollowListType] = useState<
    'followers' | 'following' | null
  >(null)
  const [isHoveringFollow, setIsHoveringFollow] = useState(false)

  const follow = useFollow(handle ?? '')
  const unfollow = useUnfollow(handle ?? '')

  const isOwner = currentUser?.username === profile?.username

  if (isLoading) {
    return (
      <div className="mx-auto max-w-[600px]">
        <p className="px-4 py-8 text-center text-muted-foreground">프로필을 불러오는 중...</p>
      </div>
    )
  }

  if (error || !profile) {
    return (
      <div className="mx-auto max-w-[600px]">
        <header className="sticky top-0 z-10 flex items-center gap-4 border-b border-border bg-background/65 px-4 py-2 backdrop-blur-xl">
          <button
            onClick={() => navigate(-1)}
            className="cursor-pointer rounded-full border-none bg-transparent p-2 text-lg text-foreground transition-colors hover:bg-foreground/10"
          >
            &larr;
          </button>
          <span className="text-xl font-bold">프로필</span>
        </header>
        <p className="px-4 py-8 text-center text-destructive">
          {error?.message ?? '사용자를 찾을 수 없습니다.'}
        </p>
      </div>
    )
  }

  const joinedDate = new Date(profile.createdAt).toLocaleDateString('ko-KR', {
    year: 'numeric',
    month: 'long',
  })

  function handleFollowClick() {
    if (profile?.isFollowing) {
      unfollow.mutate()
    } else {
      follow.mutate()
    }
  }

  return (
    <div className="mx-auto max-w-[600px]">
      <header className="sticky top-0 z-10 flex items-center gap-4 border-b border-border bg-background/65 px-4 py-2 backdrop-blur-xl">
        <button
          onClick={() => navigate(-1)}
          className="cursor-pointer rounded-full border-none bg-transparent p-2 text-lg text-foreground transition-colors hover:bg-foreground/10"
        >
          &larr;
        </button>
        <span className="text-xl font-bold">{profile.displayName}</span>
      </header>

      {profile.headerImageUrl ? (
        <img
          src={profile.headerImageUrl}
          alt="헤더 이미지"
          className="h-[200px] w-full object-cover"
        />
      ) : (
        <div className="h-[200px] w-full bg-muted-foreground/30" />
      )}

      <div className="relative px-4">
        <div className="-mt-10 flex items-start justify-between">
          {profile.profileImageUrl ? (
            <img
              src={profile.profileImageUrl}
              alt={profile.displayName}
              className="h-20 w-20 rounded-full border-4 border-background object-cover"
            />
          ) : (
            <div className="h-20 w-20 rounded-full border-4 border-background bg-muted-foreground/30" />
          )}
          {isOwner ? (
            <button
              onClick={() => setShowEditModal(true)}
              className="mt-12 cursor-pointer rounded-full border border-muted-foreground/50 bg-transparent px-4 py-1.5 text-sm font-bold text-foreground transition-colors hover:bg-foreground/10"
            >
              프로필 수정
            </button>
          ) : currentUser ? (
            <button
              onClick={handleFollowClick}
              onMouseEnter={() => setIsHoveringFollow(true)}
              onMouseLeave={() => setIsHoveringFollow(false)}
              className={cn(
                'mt-12 min-w-[100px] cursor-pointer rounded-full px-4 py-1.5 text-sm font-bold transition-all disabled:cursor-not-allowed disabled:opacity-50',
                profile.isFollowing
                  ? isHoveringFollow
                    ? 'border border-destructive/50 bg-transparent text-destructive hover:bg-destructive/10'
                    : 'border border-muted-foreground/50 bg-transparent text-foreground'
                  : 'border-none bg-foreground text-background hover:bg-foreground/90',
              )}
              disabled={follow.isPending || unfollow.isPending}
            >
              {profile.isFollowing
                ? isHoveringFollow
                  ? '언팔로우'
                  : '팔로잉'
                : '팔로우'}
            </button>
          ) : null}
        </div>

        <div className="mt-3 border-b border-border pb-4">
          <div className="text-xl font-bold">{profile.displayName}</div>
          <div className="text-[15px] text-muted-foreground">@{profile.username}</div>
          {profile.bio && <p className="mt-3 whitespace-pre-wrap text-[15px] leading-relaxed">{profile.bio}</p>}
          <div className="mt-3 text-sm text-muted-foreground">{joinedDate} 가입</div>
          <div className="mt-3 flex gap-5">
            <span
              className="cursor-pointer text-sm text-muted-foreground hover:underline"
              onClick={() => setFollowListType('following')}
            >
              <strong className="text-foreground">{profile.followingCount}</strong> 팔로잉
            </span>
            <span
              className="cursor-pointer text-sm text-muted-foreground hover:underline"
              onClick={() => setFollowListType('followers')}
            >
              <strong className="text-foreground">{profile.followersCount}</strong> 팔로워
            </span>
          </div>
        </div>
      </div>

      {showEditModal && currentUser && (
        <EditProfileModal
          user={currentUser}
          onClose={() => setShowEditModal(false)}
        />
      )}

      {followListType && handle && (
        <FollowListModal
          handle={handle}
          type={followListType}
          onClose={() => setFollowListType(null)}
        />
      )}
    </div>
  )
}
