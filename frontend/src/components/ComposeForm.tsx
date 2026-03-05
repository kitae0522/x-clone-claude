import { useState } from 'react'
import { useCreatePost } from '@/hooks/usePosts'
import { toast } from 'sonner'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
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
      {
        onSuccess: () => {
          setContent('')
          toast.success('게시글이 작성되었습니다.')
        },
        onError: (err) => {
          toast.error('게시글 작성에 실패했습니다.', { description: err.message })
        },
      },
    )
  }

  return (
    <form className="border-b border-border p-4" onSubmit={handleSubmit}>
      <Textarea
        className="w-full resize-none border-none bg-transparent py-2 text-lg shadow-none focus-visible:ring-0 placeholder:text-muted-foreground"
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
        <Button
          type="submit"
          className="rounded-full"
          disabled={remaining < 0 || content.trim().length === 0 || isPending}
        >
          {isPending ? 'Posting...' : 'Post'}
        </Button>
      </div>
    </form>
  )
}
