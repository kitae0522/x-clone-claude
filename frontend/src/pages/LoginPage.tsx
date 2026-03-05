import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Feather } from "lucide-react";
import { useLogin } from "@/hooks/useAuth";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";

export default function LoginPage() {
  const navigate = useNavigate();
  const login = useLogin();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    login.mutate({ email, password }, { onSuccess: () => navigate("/") });
  }

  return (
    <div className="flex min-h-dvh items-center justify-center p-5">
      <div className="w-full max-w-[400px]">
        <div className="mb-8 flex justify-center">
          <Feather className="h-10 w-10 text-primary" />
        </div>
        <h1 className="mb-8 text-center text-[28px] font-extrabold tracking-tight">
          로그인
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
            <Label htmlFor="password">비밀번호</Label>
            <Input
              id="password"
              type="password"
              placeholder="비밀번호를 입력하세요"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          {login.error && (
            <p className="text-[13px] text-destructive">
              {login.error.message}
            </p>
          )}
          <Button
            type="submit"
            className="mt-2 rounded-full py-6 text-[15px] font-bold"
            disabled={login.isPending}
          >
            {login.isPending ? "로그인 중..." : "로그인"}
          </Button>
        </form>
        <p className="mt-6 text-center text-sm text-muted-foreground">
          계정이 없나요?{" "}
          <Link
            to="/register"
            className="text-primary no-underline hover:underline"
          >
            회원가입
          </Link>
        </p>
      </div>
    </div>
  );
}
