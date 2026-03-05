import { Avatar, AvatarImage, AvatarFallback, type AvatarSize } from '@/components/ui/avatar'
import { cn } from '@/lib/utils'

interface UserAvatarProps {
  profileImageUrl?: string
  displayName?: string
  size?: AvatarSize
  className?: string
}

export default function UserAvatar({ profileImageUrl, displayName, size = 'md', className }: UserAvatarProps) {
  const fallbackInitial = displayName ? displayName.charAt(0).toUpperCase() : '?'

  return (
    <Avatar size={size} className={cn(className)}>
      {profileImageUrl && <AvatarImage src={profileImageUrl} alt={displayName ?? ''} />}
      <AvatarFallback>{fallbackInitial}</AvatarFallback>
    </Avatar>
  )
}
