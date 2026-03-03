import { useState } from 'react'
import { useCreatePost } from '@/hooks/usePosts'
import styles from './ComposeForm.module.css'

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
    <form className={styles.form} onSubmit={handleSubmit}>
      <textarea
        className={styles.textarea}
        placeholder="What is happening?!"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        maxLength={300}
        rows={3}
      />
      <div className={styles.footer}>
        <span
          className={`${styles.counter} ${remaining < 0 ? styles.over : remaining <= 20 ? styles.warn : ''}`}
        >
          {remaining}
        </span>
        <button
          type="submit"
          className={styles.button}
          disabled={remaining < 0 || content.trim().length === 0 || isPending}
        >
          {isPending ? 'Posting...' : 'Post'}
        </button>
      </div>
    </form>
  )
}
