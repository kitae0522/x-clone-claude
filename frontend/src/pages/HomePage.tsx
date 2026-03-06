import { usePosts } from "@/hooks/usePosts";
import PostCard from "@/components/PostCard";

export default function HomePage() {
  const { data: posts, isLoading, error } = usePosts();

  return (
    <>
      <header className="sticky top-0 z-10 border-b border-border bg-background/65 backdrop-blur-xl">
        <h1 className="px-4 py-3 text-xl font-bold">홈</h1>
      </header>
      <main>
        {isLoading && (
          <div className="flex justify-center py-8">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          </div>
        )}
        {error && (
          <p className="px-4 py-8 text-center text-destructive">
            Error: {error.message}
          </p>
        )}
        {!isLoading && !error && (!posts || posts.length === 0) && (
          <p className="px-4 py-8 text-center text-muted-foreground">
            아직 게시글이 없습니다.
          </p>
        )}
        {posts?.map((post) => (
          <PostCard key={post.id} post={post} />
        ))}
      </main>
    </>
  );
}
