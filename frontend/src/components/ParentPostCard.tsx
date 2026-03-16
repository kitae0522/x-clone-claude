import { useNavigate } from "react-router-dom";
import type { PostDetail } from "@/types/api";
import MarkdownRenderer from "@/components/MarkdownRenderer";
import MediaGrid from "@/components/MediaGrid";
import PollDisplay from "@/components/PollDisplay";
import { cn } from "@/lib/utils";

interface ParentPostCardProps {
  post: PostDetail;
}

export default function ParentPostCard({ post }: ParentPostCardProps) {
  const navigate = useNavigate();

  return (
    <div
      className="relative cursor-pointer px-4 py-3 transition-colors hover:bg-muted/30"
      onClick={() => navigate(`/post/${post.id}`)}
    >
      <div className="flex gap-3">
        <div className="flex flex-col items-center">
          <div
            className={cn(
              "shrink-0",
              !post.author.isDeleted && "cursor-pointer",
            )}
            onClick={(e) => {
              e.stopPropagation();
              if (!post.author.isDeleted) navigate(`/${post.author.username}`);
            }}
          >
            {post.author.profileImageUrl ? (
              <img
                src={post.author.profileImageUrl}
                alt=""
                className="h-10 w-10 rounded-full object-cover"
              />
            ) : (
              <div className="h-10 w-10 rounded-full bg-border" />
            )}
          </div>
          <div className="mt-1 w-0.5 flex-1 bg-border" />
        </div>
        <div className="flex-1 pb-2">
          <div className="mb-0.5 flex items-center gap-1.5">
            <span
              className={cn(
                "text-[14px] font-bold",
                post.author.isDeleted
                  ? "text-muted-foreground"
                  : "text-foreground cursor-pointer hover:underline",
              )}
              onClick={(e) => {
                e.stopPropagation();
                if (!post.author.isDeleted)
                  navigate(`/${post.author.username}`);
              }}
            >
              {post.author.displayName || post.author.username}
            </span>
            <span className="text-[13px] text-muted-foreground">
              @{post.author.username}
            </span>
            <span className="text-[13px] text-muted-foreground">
              · {new Date(post.createdAt).toLocaleString()}
            </span>
          </div>
          <div className="text-[15px] leading-normal">
            <MarkdownRenderer content={post.content} />
          </div>
          {post.media && post.media.length > 0 && (
            <MediaGrid media={post.media} />
          )}
          {post.poll && <PollDisplay poll={post.poll} postId={post.id} />}
        </div>
      </div>
    </div>
  );
}
