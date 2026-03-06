import { useRef } from "react";
import { useUpdateProfile, useUploadProfileImage } from "@/hooks/useProfile";
import { toast } from "sonner";
import type { User } from "@/types/api";
import { Camera, X as XIcon, Loader2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import UserAvatar from "@/components/UserAvatar";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useState } from "react";

interface Props {
  user: User;
  onClose: () => void;
}

const IMAGE_TYPES = ["image/jpeg", "image/png", "image/webp"];
const MAX_IMAGE_SIZE = 5 * 1024 * 1024;

function validateImageFile(file: File): string | null {
  if (!IMAGE_TYPES.includes(file.type))
    return "이미지 파일만 업로드할 수 있습니다. (JPEG, PNG, WebP)";
  if (file.size > MAX_IMAGE_SIZE) return "이미지 크기는 5MB 이하여야 합니다.";
  return null;
}

export default function EditProfileModal({ user, onClose }: Props) {
  const updateProfile = useUpdateProfile();
  const uploadImage = useUploadProfileImage();

  const [displayName, setDisplayName] = useState(user.displayName);
  const [bio, setBio] = useState(user.bio);
  const [username, setUsername] = useState(user.username);
  const [profileImageUrl, setProfileImageUrl] = useState(user.profileImageUrl);
  const [headerImageUrl, setHeaderImageUrl] = useState(user.headerImageUrl);

  const profileInputRef = useRef<HTMLInputElement>(null);
  const headerInputRef = useRef<HTMLInputElement>(null);

  function handleFileSelect(file: File, setUrl: (url: string) => void) {
    const error = validateImageFile(file);
    if (error) {
      toast.error(error);
      return;
    }
    uploadImage.mutate(file, {
      onSuccess: (url) => setUrl(url),
      onError: (err) =>
        toast.error(err instanceof Error ? err.message : "업로드 실패"),
    });
  }

  function handleSave(e: React.FormEvent) {
    e.preventDefault();
    updateProfile.mutate(
      { displayName, bio, username, profileImageUrl, headerImageUrl },
      {
        onSuccess: () => {
          toast.success("프로필이 수정되었습니다.");
          onClose();
        },
        onError: (err) => {
          toast.error("프로필 수정에 실패했습니다.", {
            description: err.message,
          });
        },
      },
    );
  }

  const isUploading = uploadImage.isPending;

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DialogContent className="max-w-[600px] p-0 [&>button:last-child]:hidden">
        <DialogHeader className="flex-row items-center gap-3 border-b border-border px-4 py-3">
          <button
            type="button"
            onClick={onClose}
            className="cursor-pointer rounded-full border-none bg-transparent p-1 text-foreground transition-colors hover:bg-foreground/10"
          >
            <XIcon className="h-5 w-5" />
          </button>
          <DialogTitle className="flex-1 text-xl">프로필 수정</DialogTitle>
          <Button
            type="button"
            onClick={handleSave}
            className="rounded-full"
            size="sm"
            disabled={updateProfile.isPending || isUploading}
          >
            {updateProfile.isPending ? "저장 중..." : "저장"}
          </Button>
        </DialogHeader>

        {updateProfile.error && (
          <p className="px-4 text-[13px] text-destructive">
            {updateProfile.error.message}
          </p>
        )}

        {/* Header Image Upload */}
        <div className="relative">
          <input
            ref={headerInputRef}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) handleFileSelect(file, setHeaderImageUrl);
              e.target.value = "";
            }}
          />
          <div
            className="relative h-[130px] w-full cursor-pointer"
            onClick={() => headerInputRef.current?.click()}
          >
            {headerImageUrl ? (
              <img
                src={headerImageUrl}
                alt="헤더"
                className="h-full w-full object-cover"
              />
            ) : (
              <div className="h-full w-full bg-muted-foreground/20" />
            )}
            <div
              className={`absolute inset-0 flex items-center justify-center transition-opacity ${
                isUploading
                  ? "bg-black/40 opacity-100"
                  : "bg-black/30 opacity-0 hover:opacity-100"
              }`}
            >
              {isUploading ? (
                <Loader2 className="h-8 w-8 animate-spin text-white" />
              ) : (
                <Camera className="h-8 w-8 text-white" />
              )}
            </div>
          </div>
        </div>

        {/* Profile Image Upload */}
        <div className="-mt-10 px-4">
          <input
            ref={profileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) handleFileSelect(file, setProfileImageUrl);
              e.target.value = "";
            }}
          />
          <div
            className="relative inline-block cursor-pointer"
            onClick={() => profileInputRef.current?.click()}
          >
            <UserAvatar
              profileImageUrl={profileImageUrl}
              displayName={displayName}
              size="xl"
              className="border-4 border-background"
            />
            <div
              className={`absolute inset-0 flex items-center justify-center rounded-full transition-opacity ${
                isUploading
                  ? "bg-black/40 opacity-100"
                  : "bg-black/30 opacity-0 hover:opacity-100"
              }`}
            >
              {isUploading ? (
                <Loader2 className="h-6 w-6 animate-spin text-white" />
              ) : (
                <Camera className="h-6 w-6 text-white" />
              )}
            </div>
          </div>
        </div>

        <form className="flex flex-col gap-4 p-4 pt-2">
          <div className="flex flex-col gap-2">
            <Label htmlFor="displayName">이름</Label>
            <Input
              id="displayName"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              maxLength={50}
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="bio">자기소개</Label>
            <Textarea
              id="bio"
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              maxLength={160}
              className="min-h-[80px] resize-y"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="username">사용자 이름</Label>
            <Input
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              maxLength={30}
            />
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
