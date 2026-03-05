import { useState } from "react";
import { useCreatePost } from "@/hooks/usePosts";
import { useAuth } from "@/hooks/useAuthContext";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import UserAvatar from "@/components/UserAvatar";

const MAX_LENGTH = 280;
const WARN_THRESHOLD = 20;
const CIRCLE_RADIUS = 10;
const CIRCLE_CIRCUMFERENCE = 2 * Math.PI * CIRCLE_RADIUS;

export default function ComposeForm() {
  const [content, setContent] = useState("");
  const { mutate, isPending } = useCreatePost();
  const { user } = useAuth();

  const charCount = [...content].length;
  const remaining = MAX_LENGTH - charCount;
  const progress = Math.min(charCount / MAX_LENGTH, 1);
  const strokeDashoffset = CIRCLE_CIRCUMFERENCE * (1 - progress);

  const circleColor =
    remaining < 0
      ? "text-destructive"
      : remaining <= WARN_THRESHOLD
        ? "text-warning"
        : "text-primary";

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (remaining < 0 || content.trim().length === 0 || isPending) return;

    mutate(
      { content, visibility: "public" },
      {
        onSuccess: () => {
          setContent("");
          toast.success("게시글이 작성되었습니다.");
        },
        onError: (err) => {
          toast.error("게시글 작성에 실패했습니다.", {
            description: err.message,
          });
        },
      },
    );
  }

  return (
    <form
      className="flex gap-3 border-b border-border p-4"
      onSubmit={handleSubmit}
    >
      <UserAvatar
        profileImageUrl={user?.profileImageUrl}
        displayName={user?.displayName}
        size="md"
        className="mt-1 shrink-0"
      />
      <div className="flex-1">
        <Textarea
          className="w-full resize-none border-none bg-transparent py-2 text-lg shadow-none focus-visible:ring-0 placeholder:text-muted-foreground"
          placeholder="무슨 일이 일어나고 있나요?"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          maxLength={300}
          rows={3}
        />
        <div className="flex items-center justify-end gap-3 border-t border-border pt-2">
          {charCount > 0 && (
            <div className="flex items-center gap-1.5">
              <svg className="h-[26px] w-[26px] -rotate-90" viewBox="0 0 24 24">
                <circle
                  cx="12"
                  cy="12"
                  r={CIRCLE_RADIUS}
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  className="text-border"
                />
                <circle
                  cx="12"
                  cy="12"
                  r={CIRCLE_RADIUS}
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeDasharray={CIRCLE_CIRCUMFERENCE}
                  strokeDashoffset={strokeDashoffset}
                  strokeLinecap="round"
                  className={`transition-all duration-200 ${circleColor}`}
                />
              </svg>
              {remaining <= WARN_THRESHOLD && (
                <span
                  className={`text-[13px] font-medium ${remaining < 0 ? "text-destructive" : "text-warning"}`}
                >
                  {remaining}
                </span>
              )}
            </div>
          )}
          <Button
            type="submit"
            className="rounded-full px-5"
            size="sm"
            disabled={remaining < 0 || content.trim().length === 0 || isPending}
          >
            {isPending ? "게시 중..." : "게시하기"}
          </Button>
        </div>
      </div>
    </form>
  );
}
