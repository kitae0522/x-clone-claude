import type { APIResponse } from "@/types/api";

const POLL_INTERVAL = 1500;
const MAX_POLL_ATTEMPTS = 120;

export interface MediaStatusResponse {
  id: string;
  status: string;
  mediaType: string;
  mimeType: string;
  width: number;
  height: number;
  size: number;
  url: string;
  error?: string;
}

export async function pollMediaStatus(
  mediaId: string,
  maxAttempts = MAX_POLL_ATTEMPTS,
): Promise<MediaStatusResponse> {
  for (let i = 0; i < maxAttempts; i++) {
    const res = await fetch(`/media/${mediaId}/status`, {
      credentials: "include",
    });
    if (!res.ok) throw new Error(`Status check failed: ${res.status}`);

    const json: APIResponse<MediaStatusResponse> = await res.json();
    if (!json.success) throw new Error(json.error ?? "Status check failed");

    if (json.data.status === "ready") {
      return json.data;
    }

    if (json.data.status === "failed") {
      throw new Error(json.data.error || "Processing failed");
    }

    await new Promise((r) => setTimeout(r, POLL_INTERVAL));
  }

  throw new Error("Processing timed out");
}
