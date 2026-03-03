import type { Post } from '@/types/api'
import styles from './PostCard.module.css'

interface PostCardProps {
  post: Post
}

const visibilityLabel: Record<Post['visibility'], string> = {
  public: 'Public',
  friends: 'Friends',
  private: 'Private',
}

function PostCard({ post }: PostCardProps) {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.author}>{post.authorId.slice(0, 8)}</span>
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
