import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useProfile } from '@/hooks/useProfile'
import { useAuth } from '@/hooks/useAuthContext'
import EditProfileModal from '@/components/EditProfileModal'
import styles from './ProfilePage.module.css'

export default function ProfilePage() {
  const { handle } = useParams<{ handle: string }>()
  const navigate = useNavigate()
  const { user: currentUser } = useAuth()
  const { data: profile, isLoading, error } = useProfile(handle ?? '')
  const [showEditModal, setShowEditModal] = useState(false)

  const isOwner = currentUser?.username === profile?.username

  if (isLoading) {
    return (
      <div className={styles.container}>
        <p className={styles.loading}>프로필을 불러오는 중...</p>
      </div>
    )
  }

  if (error || !profile) {
    return (
      <div className={styles.container}>
        <header className={styles.backHeader}>
          <button onClick={() => navigate(-1)} className={styles.backButton}>
            &larr;
          </button>
          <span className={styles.headerName}>프로필</span>
        </header>
        <p className={styles.error}>
          {error?.message ?? '사용자를 찾을 수 없습니다.'}
        </p>
      </div>
    )
  }

  const joinedDate = new Date(profile.createdAt).toLocaleDateString('ko-KR', {
    year: 'numeric',
    month: 'long',
  })

  return (
    <div className={styles.container}>
      <header className={styles.backHeader}>
        <button onClick={() => navigate(-1)} className={styles.backButton}>
          &larr;
        </button>
        <span className={styles.headerName}>{profile.displayName}</span>
      </header>

      {profile.headerImageUrl ? (
        <img
          src={profile.headerImageUrl}
          alt="헤더 이미지"
          className={styles.headerImage}
        />
      ) : (
        <div className={styles.headerImage} />
      )}

      <div className={styles.profileSection}>
        <div className={styles.avatarRow}>
          {profile.profileImageUrl ? (
            <img
              src={profile.profileImageUrl}
              alt={profile.displayName}
              className={styles.avatar}
            />
          ) : (
            <div className={styles.avatar} />
          )}
          {isOwner && (
            <button
              onClick={() => setShowEditModal(true)}
              className={styles.editButton}
            >
              프로필 수정
            </button>
          )}
        </div>

        <div className={styles.profileInfo}>
          <div className={styles.displayName}>{profile.displayName}</div>
          <div className={styles.handle}>@{profile.username}</div>
          {profile.bio && <p className={styles.bio}>{profile.bio}</p>}
          <div className={styles.joinedDate}>{joinedDate} 가입</div>
        </div>
      </div>

      {showEditModal && currentUser && (
        <EditProfileModal
          user={currentUser}
          onClose={() => setShowEditModal(false)}
        />
      )}
    </div>
  )
}
