import { usePosts } from '@/hooks/usePosts'
import { useAuth } from '@/hooks/useAuthContext'
import PostCard from '@/components/PostCard'
import styles from './HomePage.module.css'

export default function HomePage() {
  const { user, logout } = useAuth()
  const { data: posts, isLoading, error } = usePosts()

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <h1 className={styles.title}>Home</h1>
        <div className={styles.userInfo}>
          <span className={styles.username}>@{user?.username}</span>
          <button onClick={logout} className={styles.logoutButton}>
            로그아웃
          </button>
        </div>
      </header>
      <main>
        {isLoading && <p>Loading posts...</p>}
        {error && <p>Error: {error.message}</p>}
        {!isLoading && !error && (!posts || posts.length === 0) && (
          <p className={styles.empty}>No posts yet.</p>
        )}
        {posts?.map((post) => <PostCard key={post.id} post={post} />)}
      </main>
    </div>
  )
}
