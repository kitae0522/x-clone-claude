import { useMutation } from "@tanstack/react-query";
import type { APIResponse } from "@/types/api";
import { apiFetch } from "@/lib/api";

interface ChangePasswordData {
  currentPassword: string;
  newPassword: string;
}

interface DeleteAccountData {
  password: string;
}

async function changePassword(data: ChangePasswordData): Promise<void> {
  const res = await apiFetch("/api/users/password", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  const json: APIResponse<null> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to change password");
  }
}

async function deleteAccount(data: DeleteAccountData): Promise<void> {
  const res = await apiFetch("/api/users/account", {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  const json: APIResponse<null> = await res.json();
  if (!json.success) {
    throw new Error(json.error ?? "Failed to delete account");
  }
}

export function useChangePassword() {
  return useMutation({
    mutationFn: changePassword,
  });
}

export function useDeleteAccount() {
  return useMutation({
    mutationFn: deleteAccount,
  });
}
