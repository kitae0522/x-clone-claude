import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { usePosts } from '@/hooks/usePosts'
import PostCard from '@/components/PostCard'
import './App.css'

const queryClient = new QueryClient()

function PostList() {
  const { data: posts, isLoading, error } = usePosts()

  if (isLoading) {
    return <p>Loading posts...</p>
  }

  if (error) {
    return <p>Error: {error.message}</p>
  }

  if (!posts || posts.length === 0) {
    return <p>No posts yet.</p>
  }

  return (
    <div>
      {posts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}
    </div>
  )
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <div style={{ maxWidth: 600, margin: '0 auto', padding: '20px' }}>
        <h1>Posts</h1>
        <PostList />
      </div>
    </QueryClientProvider>
  )
}

export default App
