import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Feather } from 'lucide-react'
import { useUpdateProfile } from '@/hooks/useProfile'
import { useAuth } from '@/hooks/useAuthContext'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'

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

  return (
    <div className="flex min-h-dvh items-center justify-center p-5">
      <div className="w-full max-w-[400px]">
        <div className="mb-8 flex justify-center">
          <Feather className="h-10 w-10 text-primary" />
        </div>
        <h1 className="mb-2 text-center text-[28px] font-extrabold tracking-tight">환영합니다!</h1>
        <p className="mb-8 text-center text-sm text-muted-foreground">표시할 이름을 설정해주세요.</p>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="displayName">표시 이름</Label>
            <Input
              id="displayName"
              type="text"
              placeholder="이름을 입력하세요"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              maxLength={50}
              autoFocus
            />
          </div>
          {updateProfile.error && (
            <p className="text-[13px] text-destructive">{updateProfile.error.message}</p>
          )}
          <Button
            type="submit"
            className="mt-2 py-6 text-[15px] font-bold"
            disabled={updateProfile.isPending || !displayName.trim()}
          >
            {updateProfile.isPending ? '저장 중...' : '시작하기'}
          </Button>
          <Button
            type="button"
            onClick={() => navigate('/')}
            variant="outline"
            className="py-6 text-[15px] font-bold"
          >
            나중에 하기
          </Button>
        </form>
      </div>
    </div>
  )
}
