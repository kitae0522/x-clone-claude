import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useUpdateProfile } from '@/hooks/useProfile'
import { useAuth } from '@/hooks/useAuthContext'

export default function OnboardingPage() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const updateProfile = useUpdateProfile()
  const [displayName, setDisplayName] = useState('')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!displayName.trim() || !user) return

    updateProfile.mutate(
      {
        displayName: displayName.trim(),
        bio: user.bio,
        username: user.username,
        profileImageUrl: user.profileImageUrl,
        headerImageUrl: user.headerImageUrl,
      },
      { onSuccess: () => navigate('/') },
    )
  }

  function handleSkip() {
    navigate('/')
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-5">
      <div className="w-full max-w-[400px] rounded-2xl border border-border bg-background p-8">
        <h1 className="mb-2 text-center text-2xl font-bold">환영합니다!</h1>
        <p className="mb-6 text-center text-sm text-muted-foreground">표시할 이름을 설정해주세요.</p>
        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <input
            type="text"
            placeholder="표시 이름"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className="rounded-lg border border-border bg-transparent px-4 py-3 text-[15px] text-foreground outline-none transition-colors focus:border-primary"
            maxLength={50}
            autoFocus
          />
          {updateProfile.error && (
            <p className="m-0 text-[13px] text-destructive">{updateProfile.error.message}</p>
          )}
          <button
            type="submit"
            className="cursor-pointer rounded-full bg-primary py-3 text-[15px] font-bold text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
            disabled={updateProfile.isPending || !displayName.trim()}
          >
            {updateProfile.isPending ? '저장 중...' : '시작하기'}
          </button>
          <button
            type="button"
            onClick={handleSkip}
            className="cursor-pointer rounded-full border border-muted-foreground/50 bg-transparent py-3 text-[15px] font-bold text-foreground transition-colors hover:bg-foreground/10"
          >
            나중에 하기
          </button>
        </form>
      </div>
    </div>
  )
}
