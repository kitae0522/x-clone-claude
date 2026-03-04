import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useRegister } from '@/hooks/useAuth'

export default function RegisterPage() {
  const navigate = useNavigate()
  const register = useRegister()

  const [email, setEmail] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [validationError, setValidationError] = useState('')

  function validate(): boolean {
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setValidationError('올바른 이메일 형식을 입력해주세요.')
      return false
    }
    if (username.length < 3) {
      setValidationError('사용자 이름은 3자 이상이어야 합니다.')
      return false
    }
    if (password.length < 8) {
      setValidationError('비밀번호는 8자 이상이어야 합니다.')
      return false
    }
    if (password !== confirmPassword) {
      setValidationError('비밀번호가 일치하지 않습니다.')
      return false
    }
    setValidationError('')
    return true
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!validate()) return

    register.mutate(
      { email, username, password },
      { onSuccess: () => navigate('/onboarding') },
    )
  }

  const error = validationError || (register.error ? register.error.message : '')

  return (
    <div className="flex min-h-screen items-center justify-center p-5">
      <div className="w-full max-w-[400px] rounded-2xl border border-border bg-background p-8">
        <h1 className="mb-6 text-center text-2xl font-bold">회원가입</h1>
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
            type="text"
            placeholder="사용자 이름"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
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
          <input
            type="password"
            placeholder="비밀번호 확인"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            className="rounded-lg border border-border bg-transparent px-4 py-3 text-[15px] text-foreground outline-none transition-colors focus:border-primary"
            required
          />
          {error && <p className="m-0 text-[13px] text-destructive">{error}</p>}
          <button
            type="submit"
            className="cursor-pointer rounded-full bg-primary py-3 text-[15px] font-bold text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
            disabled={register.isPending}
          >
            {register.isPending ? '가입 중...' : '가입하기'}
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-muted-foreground">
          이미 계정이 있나요?{' '}
          <Link to="/login" className="text-primary no-underline hover:underline">
            로그인
          </Link>
        </p>
      </div>
    </div>
  )
}
