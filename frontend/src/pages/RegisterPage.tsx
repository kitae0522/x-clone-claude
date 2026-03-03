import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useRegister } from '@/hooks/useAuth'
import styles from './RegisterPage.module.css'

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
      { onSuccess: () => navigate('/') },
    )
  }

  const error = validationError || (register.error ? register.error.message : '')

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h1 className={styles.title}>회원가입</h1>
        <form onSubmit={handleSubmit} className={styles.form}>
          <input
            type="email"
            placeholder="이메일"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className={styles.input}
            required
          />
          <input
            type="text"
            placeholder="사용자 이름"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className={styles.input}
            required
          />
          <input
            type="password"
            placeholder="비밀번호"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className={styles.input}
            required
          />
          <input
            type="password"
            placeholder="비밀번호 확인"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            className={styles.input}
            required
          />
          {error && <p className={styles.error}>{error}</p>}
          <button
            type="submit"
            className={styles.button}
            disabled={register.isPending}
          >
            {register.isPending ? '가입 중...' : '가입하기'}
          </button>
        </form>
        <p className={styles.link}>
          이미 계정이 있나요? <Link to="/login">로그인</Link>
        </p>
      </div>
    </div>
  )
}
