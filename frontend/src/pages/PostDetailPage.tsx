import { useParams, useNavigate } from 'react-router-dom'
import { usePostDetail } from '@/hooks/usePosts'
import styles from './PostDetailPage.module.css'

export default function PostDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: post, isLoading, error } = usePostDetail(id ?? '')

  if (isLoading) return <p className={styles.status}>Loading...</p>
  if (error) return <p className={styles.status}>Error: {error.message}</p>
  if (!post) return <p className={styles.status}>Post not found.</p>

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <button className={styles.backButton} onClick={() => navigate(-1)}>
          &larr;
        </button>
        <h1 className={styles.title}>Post</h1>
      </header>
      <article className={styles.post}>
        <div className={styles.authorRow}>
          {post.author.profileImageUrl ? (
            <img
              src={post.author.profileImageUrl}
              alt=""
              className={styles.avatar}
            />
          ) : (
            <div className={styles.avatarPlaceholder} />
          )}
          <div className={styles.authorText}>
            <span className={styles.displayName}>
              {post.author.displayName || post.author.username}
            </span>
            <span className={styles.username}>@{post.author.username}</span>
          </div>
        </div>
        <p className={styles.content}>{post.content}</p>
        <span className={styles.date}>
          {new Date(post.createdAt).toLocaleString()}
        </span>
      </article>
    </div>
  )
}
