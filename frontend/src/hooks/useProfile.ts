import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type {
  APIResponse,
  ProfileUser,
  UpdateProfileRequest,
  User,
} from "@/types/api";
import { apiFetch } from "@/lib/api";

async function fetchProfile(handle: string): Promise<ProfileUser> {
  const res = await apiFetch(`/api/users/${handle}`);
  const json: APIResponse<ProfileUser> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to fetch profile");
  }
  return json.data;
}

async function putUpdateProfile(data: UpdateProfileRequest): Promise<User> {
  const res = await apiFetch("/api/users/profile", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  const json: APIResponse<User> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to update profile");
  }
  return json.data;
}

export function useProfile(handle: string, enabled: boolean = true) {
  return useQuery({
    queryKey: ["users", handle],
    queryFn: () => fetchProfile(handle),
    enabled: !!handle && enabled,
  });
}

const POLL_INTERVAL = 1500;
const MAX_POLL_ATTEMPTS = 60;

async function postUploadImage(file: File): Promise<string> {
  const formData = new FormData();
  formData.append("file", file);

  // Step 1: Upload to media-service (S3)
  const uploadRes = await fetch("/media/upload", {
    method: "POST",
    body: formData,
    credentials: "include",
  });
  const uploadJson: APIResponse<{ id: string }> = await uploadRes.json();
  if (!uploadJson.success) {
    throw new Error(uploadJson.error ?? "이미지 업로드에 실패했습니다.");
  }

  const mediaId = uploadJson.data.id;

  // Step 2: Poll for processing completion (resize + WebP conversion)
  for (let i = 0; i < MAX_POLL_ATTEMPTS; i++) {
    const statusRes = await fetch(`/media/${mediaId}/status`, {
      credentials: "include",
    });
    if (!statusRes.ok) throw new Error("상태 확인에 실패했습니다.");

    const statusJson: APIResponse<{ status: string; error?: string }> =
      await statusRes.json();
    if (!statusJson.success)
      throw new Error(statusJson.error ?? "상태 확인에 실패했습니다.");

    if (statusJson.data.status === "ready") {
      return `/media/${mediaId}?size=medium`;
    }
    if (statusJson.data.status === "failed") {
      throw new Error(statusJson.data.error || "이미지 처리에 실패했습니다.");
    }

    await new Promise((r) => setTimeout(r, POLL_INTERVAL));
  }

  throw new Error("이미지 처리 시간이 초과되었습니다.");
}

export function useUploadProfileImage() {
  return useMutation({
    mutationFn: postUploadImage,
  });
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: putUpdateProfile,
    onSuccess: (user) => {
      queryClient.setQueryData(["auth", "me"], user);
      queryClient.invalidateQueries({ queryKey: ["users", user.username] });
    },
  });
}
