import { useNavigate } from 'react-router-dom'
import type { PostDetail } from '@/types/api'
import styles from './PostCard.module.css'

interface PostCardProps {
  post: PostDetail
}

const visibilityLabel: Record<PostDetail['visibility'], string> = {
  public: 'Public',
  friends: 'Friends',
  private: 'Private',
}

function PostCard({ post }: PostCardProps) {
  const navigate = useNavigate()

  return (
    <div
      className={styles.card}
      onClick={() => navigate(`/post/${post.id}`)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter') navigate(`/post/${post.id}`)
      }}
    >
      <div className={styles.header}>
        <div className={styles.authorInfo}>
          {post.author.profileImageUrl ? (
            <img
              src={post.author.profileImageUrl}
              alt=""
              className={styles.avatar}
            />
          ) : (
            <div className={styles.avatarPlaceholder} />
          )}
          <div>
            <span className={styles.displayName}>
              {post.author.displayName || post.author.username}
            </span>
            <span className={styles.username}>@{post.author.username}</span>
          </div>
        </div>
        <span className={`${styles.badge} ${styles[post.visibility]}`}>
          {visibilityLabel[post.visibility]}
        </span>
      </div>
      <p className={styles.content}>{post.content}</p>
      <span className={styles.date}>
        {new Date(post.createdAt).toLocaleString()}
      </span>
    </div>
  )
}

export default PostCard
