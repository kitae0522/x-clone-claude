import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useUpdateProfile } from '@/hooks/useProfile'
import { useAuth } from '@/hooks/useAuthContext'
import styles from './OnboardingPage.module.css'

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
    <div className={styles.container}>
      <div className={styles.card}>
        <h1 className={styles.title}>환영합니다!</h1>
        <p className={styles.subtitle}>표시할 이름을 설정해주세요.</p>
        <form onSubmit={handleSubmit} className={styles.form}>
          <input
            type="text"
            placeholder="표시 이름"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className={styles.input}
            maxLength={50}
            autoFocus
          />
          {updateProfile.error && (
            <p className={styles.error}>{updateProfile.error.message}</p>
          )}
          <button
            type="submit"
            className={styles.button}
            disabled={updateProfile.isPending || !displayName.trim()}
          >
            {updateProfile.isPending ? '저장 중...' : '시작하기'}
          </button>
          <button
            type="button"
            onClick={handleSkip}
            className={styles.skipButton}
          >
            나중에 하기
          </button>
        </form>
      </div>
    </div>
  )
}
