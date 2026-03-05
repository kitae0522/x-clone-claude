import { useState } from "react";
import { useCreateReply } from "@/hooks/useReplies";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const MAX_LENGTH = 280;

interface ReplyFormProps {
  postId: string;
  parentPostId?: string;
}

export default function ReplyForm({ postId, parentPostId }: ReplyFormProps) {
  const [content, setContent] = useState("");
  const { mutate, isPending } = useCreateReply(postId, parentPostId);

  const remaining = MAX_LENGTH - [...content].length;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (remaining < 0 || content.trim().length === 0 || isPending) return;

    mutate(
      { content },
      {
        onSuccess: () => {
          setContent("");
          toast.success("댓글이 작성되었습니다.");
        },
        onError: (err) => {
          toast.error("댓글 작성에 실패했습니다.", { description: err.message });
        },
      },
    );
  }

  return (
    <form className="border-b border-border p-4" onSubmit={handleSubmit}>
      <Textarea
        className="w-full resize-none border-none bg-transparent py-2 text-[15px] shadow-none focus-visible:ring-0 placeholder:text-muted-foreground"
        placeholder="Post your reply"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        maxLength={MAX_LENGTH}
        rows={2}
      />
      <div className="flex items-center justify-end gap-3 border-t border-border pt-2">
        <span
          className={cn(
            "text-sm text-muted-foreground",
            remaining < 0 && "text-destructive",
            remaining >= 0 && remaining <= 20 && "text-warning",
          )}
        >
          {remaining}
        </span>
        <Button
          type="submit"
          size="sm"
          className="rounded-full"
          disabled={remaining < 0 || content.trim().length === 0 || isPending}
        >
          {isPending ? "Replying..." : "Reply"}
        </Button>
      </div>
    </form>
  );
}
