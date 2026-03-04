import { useState } from 'react'
import { useCreatePost } from '@/hooks/usePosts'
import { cn } from '@/lib/utils'

const MAX_LENGTH = 280

export default function ComposeForm() {
  const [content, setContent] = useState('')
  const { mutate, isPending } = useCreatePost()

  const remaining = MAX_LENGTH - [...content].length

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (remaining < 0 || content.trim().length === 0 || isPending) return

    mutate(
      { content, visibility: 'public' },
      { onSuccess: () => setContent('') },
    )
  }

  return (
    <form className="border-b border-border p-4" onSubmit={handleSubmit}>
      <textarea
        className="w-full resize-none border-none bg-transparent py-2 font-[inherit] text-lg text-foreground outline-none placeholder:text-muted-foreground"
        placeholder="What is happening?!"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        maxLength={300}
        rows={3}
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
          className="cursor-pointer rounded-full bg-primary px-5 py-2 text-[15px] font-bold text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
          disabled={remaining < 0 || content.trim().length === 0 || isPending}
        >
          {isPending ? 'Posting...' : 'Post'}
        </button>
      </div>
    </form>
  )
}
