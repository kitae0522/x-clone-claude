import { useNavigate } from 'react-router-dom'
import { useFollowers, useFollowing } from '@/hooks/useFollow'
import UserAvatar from '@/components/UserAvatar'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface Props {
  handle: string
  type: 'followers' | 'following'
  onClose: () => void
}

export default function FollowListModal({ handle, type, onClose }: Props) {
  const navigate = useNavigate()
  const { data: followingData, isLoading: followingLoading } = useFollowing(
    handle,
    type === 'following',
  )
  const { data: followersData, isLoading: followersLoading } = useFollowers(
    handle,
    type === 'followers',
  )

  const data = type === 'following' ? followingData : followersData
  const isLoading = type === 'following' ? followingLoading : followersLoading
  const title = type === 'following' ? '팔로잉' : '팔로워'

  function handleUserClick(username: string) {
    onClose()
    navigate(`/${username}`)
  }

  return (
    <Dialog open onOpenChange={(open) => { if (!open) onClose() }}>
      <DialogContent className="max-h-[90vh] max-w-[600px] overflow-y-auto p-0">
        <DialogHeader className="sticky top-0 z-[1] bg-background/65 px-4 py-3 backdrop-blur-xl">
          <DialogTitle className="text-xl">{title}</DialogTitle>
        </DialogHeader>

        {isLoading ? (
          <p className="px-4 py-8 text-center text-muted-foreground">불러오는 중...</p>
        ) : !data || data.users.length === 0 ? (
          <p className="px-4 py-8 text-center text-muted-foreground">
            {type === 'following'
              ? '아직 팔로우하는 사용자가 없습니다.'
              : '아직 팔로워가 없습니다.'}
          </p>
        ) : (
          <div>
            {data.users.map((user) => (
              <div
                key={user.id}
                className="flex cursor-pointer items-start gap-3 px-4 py-3 transition-colors hover:bg-foreground/[0.03]"
                onClick={() => handleUserClick(user.username)}
              >
                <UserAvatar
                  profileImageUrl={user.profileImageUrl}
                  displayName={user.displayName}
                  size="md"
                />
                <div className="min-w-0 flex-1">
                  <div className="text-[15px] font-bold">{user.displayName}</div>
                  <div className="text-[15px] text-muted-foreground">@{user.username}</div>
                  {user.bio && <p className="mt-1 whitespace-pre-wrap text-[15px] leading-relaxed">{user.bio}</p>}
                </div>
              </div>
            ))}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
