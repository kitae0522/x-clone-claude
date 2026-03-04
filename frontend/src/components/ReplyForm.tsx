import { useState } from 'react'
import { useCreateReply } from '@/hooks/useReplies'
import { cn } from '@/lib/utils'

const MAX_LENGTH = 280

interface ReplyFormProps {
  postId: string
}

export default function ReplyForm({ postId }: ReplyFormProps) {
  const [content, setContent] = useState('')
  const { mutate, isPending } = useCreateReply(postId)

  const remaining = MAX_LENGTH - [...content].length

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (remaining < 0 || content.trim().length === 0 || isPending) return

    mutate(
      { content },
      { onSuccess: () => setContent('') },
    )
  }

  return (
    <form className="border-b border-border p-4" onSubmit={handleSubmit}>
      <textarea
        className="w-full resize-none border-none bg-transparent py-2 font-[inherit] text-[15px] text-foreground outline-none placeholder:text-muted-foreground"
        placeholder="Post your reply"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        maxLength={300}
        rows={2}
      />
      <div className="flex items-center justify-end gap-3 border-t border-border pt-2">
        <span
          className={cn(
            'text-sm text-muted-foreground',
            remaining < 0 && 'text-destructive',
            remaining >= 0 && remaining <= 20 && 'text-warning',
          )}
        >
          {remaining}
        </span>
        <button
          type="submit"
          className="cursor-pointer rounded-full bg-primary px-4 py-1.5 text-[13px] font-bold text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
          disabled={remaining < 0 || content.trim().length === 0 || isPending}
        >
          {isPending ? 'Replying...' : 'Reply'}
        </button>
      </div>
    </form>
  )
}
