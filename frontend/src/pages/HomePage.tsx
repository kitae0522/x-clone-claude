import { usePosts } from '@/hooks/usePosts'
import { useAuth } from '@/hooks/useAuthContext'
import PostCard from '@/components/PostCard'
import ComposeForm from '@/components/ComposeForm'

export default function HomePage() {
  const { user, logout } = useAuth()
  const { data: posts, isLoading, error } = usePosts()

  return (
    <div className="mx-auto max-w-[600px]">
      <header className="sticky top-0 z-10 flex items-center justify-between border-b border-border bg-background/65 px-4 py-3 backdrop-blur-xl">
        <h1 className="text-xl font-bold">Home</h1>
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">@{user?.username}</span>
          <button
            onClick={logout}
            className="cursor-pointer rounded-full border border-muted-foreground/50 bg-transparent px-4 py-1.5 text-sm font-bold text-foreground transition-colors hover:bg-foreground/10"
          >
            로그아웃
          </button>
        </div>
      </header>
      <ComposeForm />
      <main>
        {isLoading && <p>Loading posts...</p>}
        {error && <p>Error: {error.message}</p>}
        {!isLoading && !error && (!posts || posts.length === 0) && (
          <p className="px-4 py-8 text-center text-muted-foreground">No posts yet.</p>
        )}
        {posts?.map((post) => <PostCard key={post.id} post={post} />)}
      </main>
    </div>
  )
}
