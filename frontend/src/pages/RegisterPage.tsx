import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Feather } from "lucide-react";
import { useRegister } from "@/hooks/useAuth";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";

export default function RegisterPage() {
  const navigate = useNavigate();
  const register = useRegister();

  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [validationError, setValidationError] = useState("");

  function validate(): boolean {
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setValidationError("올바른 이메일 형식을 입력해주세요.");
      return false;
    }
    if (username.length < 3) {
      setValidationError("사용자 이름은 3자 이상이어야 합니다.");
      return false;
    }
    if (password.length < 8) {
      setValidationError("비밀번호는 8자 이상이어야 합니다.");
      return false;
    }
    if (password !== confirmPassword) {
      setValidationError("비밀번호가 일치하지 않습니다.");
      return false;
    }
    setValidationError("");
    return true;
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!validate()) return;

    register.mutate(
      { email, username, password },
      { onSuccess: () => navigate("/onboarding") },
    );
  }

  const error =
    validationError || (register.error ? register.error.message : "");

  return (
    <div className="flex min-h-dvh items-center justify-center p-5">
      <div className="w-full max-w-[400px]">
        <div className="mb-8 flex justify-center">
          <Feather className="h-10 w-10 text-primary" />
        </div>
        <h1 className="mb-8 text-center text-[28px] font-extrabold tracking-tight">
          계정 만들기
        </h1>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="email">이메일</Label>
            <Input
              id="email"
              type="email"
              placeholder="name@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="username">사용자 이름</Label>
            <Input
              id="username"
              type="text"
              placeholder="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="password">비밀번호</Label>
            <Input
              id="password"
              type="password"
              placeholder="8자 이상"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="confirmPassword">비밀번호 확인</Label>
            <Input
              id="confirmPassword"
              type="password"
              placeholder="비밀번호를 다시 입력하세요"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
          </div>
          {error && <p className="text-[13px] text-destructive">{error}</p>}
          <Button
            type="submit"
            className="mt-2 rounded-full py-6 text-[15px] font-bold"
            disabled={register.isPending}
          >
            {register.isPending ? "가입 중..." : "가입하기"}
          </Button>
        </form>
        <p className="mt-6 text-center text-sm text-muted-foreground">
          이미 계정이 있나요?{" "}
          <Link
            to="/login"
            className="text-primary no-underline hover:underline"
          >
            로그인
          </Link>
        </p>
      </div>
    </div>
  );
}
