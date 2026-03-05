import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";

export default function RightPanel() {
  return (
    <aside className="sticky top-0 h-dvh px-4 py-3">
      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="검색"
          className="h-[42px] rounded-full border-none bg-secondary pl-10 text-[15px] focus-visible:ring-1 focus-visible:ring-primary"
        />
      </div>

      {/* Trending / What's happening */}
      <div className="mt-4 overflow-hidden rounded-2xl bg-secondary">
        <h2 className="px-4 py-3 text-[20px] font-extrabold">트렌드</h2>
        {[
          { category: "기술", topic: "React 19", posts: "12.4K" },
          { category: "개발", topic: "TypeScript", posts: "8.2K" },
          { category: "한국", topic: "Claude Code", posts: "5.1K" },
        ].map(({ category, topic, posts }) => (
          <div
            key={topic}
            className="cursor-pointer px-4 py-3 transition-colors hover:bg-foreground/5"
          >
            <div className="text-[13px] text-muted-foreground">
              {category}에서 트렌드
            </div>
            <div className="text-[15px] font-bold">{topic}</div>
            <div className="text-[13px] text-muted-foreground">
              {posts} posts
            </div>
          </div>
        ))}
        <div className="cursor-pointer px-4 py-3 text-[15px] text-primary transition-colors hover:bg-foreground/5">
          더 보기
        </div>
      </div>
    </aside>
  );
}
