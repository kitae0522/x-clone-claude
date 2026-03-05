import { useState } from 'react'
import { useUpdateProfile } from '@/hooks/useProfile'
import { toast } from 'sonner'
import type { User } from '@/types/api'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface Props {
  user: User
  onClose: () => void
}

export default function EditProfileModal({ user, onClose }: Props) {
  const updateProfile = useUpdateProfile()

  const [displayName, setDisplayName] = useState(user.displayName)
  const [bio, setBio] = useState(user.bio)
  const [username, setUsername] = useState(user.username)
  const [profileImageUrl, setProfileImageUrl] = useState(user.profileImageUrl)
  const [headerImageUrl, setHeaderImageUrl] = useState(user.headerImageUrl)

  function handleSave(e: React.FormEvent) {
    e.preventDefault()
    updateProfile.mutate(
      { displayName, bio, username, profileImageUrl, headerImageUrl },
      {
        onSuccess: () => {
          toast.success('프로필이 수정되었습니다.')
          onClose()
        },
        onError: (err) => {
          toast.error('프로필 수정에 실패했습니다.', { description: err.message })
        },
      },
    )
  }

  return (
    <Dialog open onOpenChange={(open) => { if (!open) onClose() }}>
      <DialogContent className="max-w-[600px] p-0">
        <DialogHeader className="flex-row items-center justify-between border-b border-border px-4 py-3">
          <DialogTitle className="text-xl">프로필 수정</DialogTitle>
          <Button
            onClick={handleSave}
            className="rounded-full"
            size="sm"
            disabled={updateProfile.isPending}
          >
            {updateProfile.isPending ? '저장 중...' : '저장'}
          </Button>
        </DialogHeader>

        {updateProfile.error && (
          <p className="px-4 text-[13px] text-destructive">{updateProfile.error.message}</p>
        )}

        <form onSubmit={handleSave} className="flex flex-col gap-4 p-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="displayName">이름</Label>
            <Input
              id="displayName"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              maxLength={50}
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="bio">자기소개</Label>
            <Textarea
              id="bio"
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              maxLength={160}
              className="min-h-[80px] resize-y"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="username">사용자 이름</Label>
            <Input
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="profileImageUrl">프로필 이미지 URL</Label>
            <Input
              id="profileImageUrl"
              type="url"
              value={profileImageUrl}
              onChange={(e) => setProfileImageUrl(e.target.value)}
              placeholder="https://example.com/avatar.jpg"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="headerImageUrl">헤더 이미지 URL</Label>
            <Input
              id="headerImageUrl"
              type="url"
              value={headerImageUrl}
              onChange={(e) => setHeaderImageUrl(e.target.value)}
              placeholder="https://example.com/header.jpg"
            />
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
