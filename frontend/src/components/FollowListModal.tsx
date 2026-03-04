import { useNavigate } from 'react-router-dom'
import { useFollowers, useFollowing } from '@/hooks/useFollow'
import styles from './FollowListModal.module.css'

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
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <button onClick={onClose} className={styles.closeButton}>
            &times;
          </button>
          <span className={styles.title}>{title}</span>
        </div>

        {isLoading ? (
          <p className={styles.loading}>불러오는 중...</p>
        ) : !data || data.users.length === 0 ? (
          <p className={styles.empty}>
            {type === 'following'
              ? '아직 팔로우하는 사용자가 없습니다.'
              : '아직 팔로워가 없습니다.'}
          </p>
        ) : (
          <div className={styles.list}>
            {data.users.map((user) => (
              <div
                key={user.id}
                className={styles.userItem}
                onClick={() => handleUserClick(user.username)}
              >
                {user.profileImageUrl ? (
                  <img
                    src={user.profileImageUrl}
                    alt={user.displayName}
                    className={styles.avatar}
                  />
                ) : (
                  <div className={styles.avatar} />
                )}
                <div className={styles.userInfo}>
                  <div className={styles.displayName}>{user.displayName}</div>
                  <div className={styles.handle}>@{user.username}</div>
                  {user.bio && <p className={styles.bio}>{user.bio}</p>}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
