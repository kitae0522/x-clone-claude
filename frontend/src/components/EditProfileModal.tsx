import { useState } from 'react'
import { useUpdateProfile } from '@/hooks/useProfile'
import type { User } from '@/types/api'
import styles from './EditProfileModal.module.css'

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
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <div className={styles.headerLeft}>
            <button onClick={onClose} className={styles.closeButton}>
              &times;
            </button>
            <span className={styles.title}>프로필 수정</span>
          </div>
          <button
            onClick={handleSave}
            className={styles.saveButton}
            disabled={updateProfile.isPending}
          >
            {updateProfile.isPending ? '저장 중...' : '저장'}
          </button>
        </div>

        {updateProfile.error && (
          <p className={styles.error}>{updateProfile.error.message}</p>
        )}

        <form onSubmit={handleSave} className={styles.form}>
          <div className={styles.field}>
            <label className={styles.label}>이름</label>
            <input
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              className={styles.input}
              maxLength={50}
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label}>자기소개</label>
            <textarea
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              className={styles.textarea}
              maxLength={160}
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label}>사용자 이름</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className={styles.input}
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label}>프로필 이미지 URL</label>
            <input
              type="url"
              value={profileImageUrl}
              onChange={(e) => setProfileImageUrl(e.target.value)}
              className={styles.input}
              placeholder="https://example.com/avatar.jpg"
            />
          </div>

          <div className={styles.field}>
            <label className={styles.label}>헤더 이미지 URL</label>
            <input
              type="url"
              value={headerImageUrl}
              onChange={(e) => setHeaderImageUrl(e.target.value)}
              className={styles.input}
              placeholder="https://example.com/header.jpg"
            />
          </div>
        </form>
      </div>
    </div>
  )
}
