import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuthContext";
import { useChangePassword, useDeleteAccount } from "@/hooks/useSettings";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";

export default function SettingsPage() {
  const navigate = useNavigate();
  const { logout } = useAuth();

  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const [deletePassword, setDeletePassword] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const changePasswordMutation = useChangePassword();
  const deleteAccountMutation = useDeleteAccount();

  const passwordError =
    confirmPassword && newPassword !== confirmPassword
      ? "비밀번호가 일치하지 않습니다"
      : newPassword && newPassword.length < 8
        ? "비밀번호는 8자 이상이어야 합니다"
        : newPassword && newPassword.length > 128
          ? "비밀번호는 128자 이하여야 합니다"
          : "";

  const canSubmitPassword =
    currentPassword &&
    newPassword.length >= 8 &&
    newPassword.length <= 128 &&
    newPassword === confirmPassword &&
    !changePasswordMutation.isPending;

  function handleChangePassword(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmitPassword) return;

    changePasswordMutation.mutate(
      { currentPassword, newPassword },
      {
        onSuccess: () => {
          toast.success("비밀번호가 변경되었습니다");
          setCurrentPassword("");
          setNewPassword("");
          setConfirmPassword("");
        },
        onError: (err) => {
          toast.error(err.message);
        },
      },
    );
  }

  function handleDeleteAccount() {
    if (!deletePassword) return;

    deleteAccountMutation.mutate(
      { password: deletePassword },
      {
        onSuccess: () => {
          toast.success("계정이 삭제되었습니다");
          setDeleteDialogOpen(false);
          logout();
          navigate("/login");
        },
        onError: (err) => {
          toast.error(err.message);
        },
      },
    );
  }

  return (
    <div className="mx-auto w-full max-w-xl px-4 py-6">
      <h1 className="mb-6 text-xl font-bold">설정</h1>

      {/* Trash Link */}
      <section className="mb-8">
        <button
          onClick={() => navigate("/trash")}
          className="flex w-full cursor-pointer items-center gap-3 rounded-lg border border-border bg-transparent px-4 py-3 text-left transition-colors hover:bg-muted"
        >
          <Trash2 className="h-5 w-5 text-muted-foreground" />
          <div>
            <p className="text-sm font-medium">휴지통</p>
            <p className="text-xs text-muted-foreground">
              삭제된 게시글을 관리합니다
            </p>
          </div>
        </button>
      </section>

      {/* Password Change Section */}
      <section className="mb-8">
        <h2 className="mb-4 text-lg font-semibold">비밀번호 변경</h2>
        <form onSubmit={handleChangePassword} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="currentPassword">현재 비밀번호</Label>
            <Input
              id="currentPassword"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              placeholder="현재 비밀번호를 입력하세요"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="newPassword">새 비밀번호</Label>
            <Input
              id="newPassword"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="새 비밀번호를 입력하세요 (8자 이상)"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="confirmPassword">비밀번호 확인</Label>
            <Input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="새 비밀번호를 다시 입력하세요"
            />
            {passwordError && (
              <p className="text-sm text-destructive">{passwordError}</p>
            )}
          </div>
          <Button type="submit" disabled={!canSubmitPassword}>
            {changePasswordMutation.isPending ? "변경 중..." : "비밀번호 변경"}
          </Button>
        </form>
      </section>

      {/* Account Deletion Section */}
      <section className="rounded-lg border border-destructive/50 p-4">
        <h2 className="mb-2 text-lg font-semibold text-destructive">
          계정 탈퇴
        </h2>
        <p className="mb-4 text-sm text-muted-foreground">
          계정을 탈퇴하면 더 이상 로그인할 수 없습니다. 이 작업은 되돌릴 수
          없습니다.
        </p>
        <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
          <AlertDialogTrigger asChild>
            <Button variant="destructive">계정 탈퇴</Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>정말 탈퇴하시겠습니까?</AlertDialogTitle>
              <AlertDialogDescription>
                이 작업은 되돌릴 수 없습니다. 계정이 영구적으로 비활성화됩니다.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <div className="space-y-2 py-2">
              <Label htmlFor="deletePassword">비밀번호 확인</Label>
              <Input
                id="deletePassword"
                type="password"
                value={deletePassword}
                onChange={(e) => setDeletePassword(e.target.value)}
                placeholder="비밀번호를 입력하세요"
              />
            </div>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setDeletePassword("")}>
                취소
              </AlertDialogCancel>
              <AlertDialogAction
                onClick={handleDeleteAccount}
                disabled={!deletePassword || deleteAccountMutation.isPending}
                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              >
                {deleteAccountMutation.isPending ? "처리 중..." : "탈퇴하기"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </section>
    </div>
  );
}
