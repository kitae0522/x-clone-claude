import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useLogin } from '@/hooks/useAuth'

export default function LoginPage() {
  const navigate = useNavigate()
  const login = useLogin()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    login.mutate(
      { email, password },
      { onSuccess: () => navigate('/') },
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-5">
      <div className="w-full max-w-[400px] rounded-2xl border border-border bg-background p-8">
        <h1 className="mb-6 text-center text-2xl font-bold">로그인</h1>
        <form onSubmit={handleSubmit} className="flex flex-col gap-3">
          <input
            type="email"
            placeholder="이메일"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="rounded-lg border border-border bg-transparent px-4 py-3 text-[15px] text-foreground outline-none transition-colors focus:border-primary"
            required
          />
          <input
            type="password"
            placeholder="비밀번호"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="rounded-lg border border-border bg-transparent px-4 py-3 text-[15px] text-foreground outline-none transition-colors focus:border-primary"
            required
          />
          {login.error && (
            <p className="m-0 text-[13px] text-destructive">{login.error.message}</p>
          )}
          <button
            type="submit"
            className="cursor-pointer rounded-full bg-primary py-3 text-[15px] font-bold text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
            disabled={login.isPending}
          >
            {login.isPending ? '로그인 중...' : '로그인'}
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-muted-foreground">
          계정이 없나요?{' '}
          <Link to="/register" className="text-primary no-underline hover:underline">
            회원가입
          </Link>
        </p>
      </div>
    </div>
  )
}
