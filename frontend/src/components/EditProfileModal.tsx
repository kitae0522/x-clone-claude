import { useState } from 'react'
import { useUpdateProfile } from '@/hooks/useProfile'
import type { User } from '@/types/api'

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
        onSuccess: () => onClose(),
      },
    )
  }

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-muted-foreground/40"
      onClick={onClose}
    >
      <div
        className="w-full max-w-[600px] max-h-[90vh] overflow-y-auto rounded-2xl bg-background"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="sticky top-0 z-[1] flex h-[53px] items-center justify-between bg-background/65 px-4 backdrop-blur-xl">
          <div className="flex items-center gap-4">
            <button
              onClick={onClose}
              className="cursor-pointer rounded-full border-none bg-transparent p-2 text-lg text-foreground transition-colors hover:bg-foreground/10"
            >
              &times;
            </button>
            <span className="text-xl font-bold">프로필 수정</span>
          </div>
          <button
            onClick={handleSave}
            className="cursor-pointer rounded-full border-none bg-foreground px-4 py-1.5 text-sm font-bold text-background transition-colors hover:bg-foreground/90 disabled:cursor-not-allowed disabled:opacity-50"
            disabled={updateProfile.isPending}
          >
            {updateProfile.isPending ? '저장 중...' : '저장'}
          </button>
        </div>

        {updateProfile.error && (
          <p className="px-4 text-[13px] text-destructive">{updateProfile.error.message}</p>
        )}

        <form onSubmit={handleSave} className="flex flex-col gap-4 p-4">
          <div className="flex flex-col gap-1">
            <label className="text-[13px] text-muted-foreground">이름</label>
            <input
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              className="rounded-lg border border-border bg-transparent px-4 py-3 text-[15px] text-foreground outline-none transition-colors focus:border-primary"
              maxLength={50}
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-[13px] text-muted-foreground">자기소개</label>
            <textarea
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              className="min-h-[80px] resize-y rounded-lg border border-border bg-transparent px-4 py-3 font-[inherit] text-[15px] text-foreground outline-none transition-colors focus:border-primary"
              maxLength={160}
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-[13px] text-muted-foreground">사용자 이름</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="rounded-lg border border-border bg-transparent px-4 py-3 text-[15px] text-foreground outline-none transition-colors focus:border-primary"
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-[13px] text-muted-foreground">프로필 이미지 URL</label>
            <input
              type="url"
              value={profileImageUrl}
              onChange={(e) => setProfileImageUrl(e.target.value)}
              className="rounded-lg border border-border bg-transparent px-4 py-3 text-[15px] text-foreground outline-none transition-colors focus:border-primary"
              placeholder="https://example.com/avatar.jpg"
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-[13px] text-muted-foreground">헤더 이미지 URL</label>
            <input
              type="url"
              value={headerImageUrl}
              onChange={(e) => setHeaderImageUrl(e.target.value)}
              className="rounded-lg border border-border bg-transparent px-4 py-3 text-[15px] text-foreground outline-none transition-colors focus:border-primary"
              placeholder="https://example.com/header.jpg"
            />
          </div>
        </form>
      </div>
    </div>
  )
}
