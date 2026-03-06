import { useState, useCallback } from "react";
import type { APIResponse, MediaItem } from "@/types/api";

interface UploadProgress {
  id: string;
  progress: number;
  status: "uploading" | "done" | "error";
  media?: MediaItem;
  error?: string;
}

const MAX_IMAGE_SIZE = 5 * 1024 * 1024;
const MAX_VIDEO_SIZE = 50 * 1024 * 1024;
const MAX_GIF_SIZE = 15 * 1024 * 1024;

const ALLOWED_TYPES: Record<string, { type: MediaItem["type"]; maxSize: number }> = {
  "image/jpeg": { type: "image", maxSize: MAX_IMAGE_SIZE },
  "image/png": { type: "image", maxSize: MAX_IMAGE_SIZE },
  "image/webp": { type: "image", maxSize: MAX_IMAGE_SIZE },
  "image/gif": { type: "gif", maxSize: MAX_GIF_SIZE },
  "video/mp4": { type: "video", maxSize: MAX_VIDEO_SIZE },
  "video/webm": { type: "video", maxSize: MAX_VIDEO_SIZE },
};

export function useMediaUpload() {
  const [uploads, setUploads] = useState<UploadProgress[]>([]);
  const [mediaItems, setMediaItems] = useState<MediaItem[]>([]);

  const isUploading = uploads.some((u) => u.status === "uploading");

  const validateFile = useCallback((file: File): string | null => {
    const config = ALLOWED_TYPES[file.type];
    if (!config) {
      return `허용되지 않은 파일 형식입니다: ${file.type}`;
    }
    if (file.size > config.maxSize) {
      const maxMB = config.maxSize / (1024 * 1024);
      return `파일 크기가 ${maxMB}MB를 초과합니다.`;
    }
    return null;
  }, []);

  const uploadFile = useCallback(async (file: File): Promise<MediaItem | null> => {
    const tempId = crypto.randomUUID();

    setUploads((prev) => [...prev, { id: tempId, progress: 0, status: "uploading" }]);

    const formData = new FormData();
    formData.append("file", file);

    try {
      const xhr = new XMLHttpRequest();

      const result = await new Promise<MediaItem>((resolve, reject) => {
        xhr.upload.addEventListener("progress", (e) => {
          if (e.lengthComputable) {
            const progress = Math.round((e.loaded / e.total) * 100);
            setUploads((prev) =>
              prev.map((u) => (u.id === tempId ? { ...u, progress } : u)),
            );
          }
        });

        xhr.addEventListener("load", () => {
          if (xhr.status >= 200 && xhr.status < 300) {
            const json: APIResponse<MediaItem> = JSON.parse(xhr.responseText);
            if (json.success) {
              resolve(json.data);
            } else {
              reject(new Error(json.error ?? "Upload failed"));
            }
          } else {
            reject(new Error(`Upload failed: ${xhr.status}`));
          }
        });

        xhr.addEventListener("error", () => reject(new Error("Network error")));
        xhr.open("POST", "/api/media/upload");
        xhr.withCredentials = true;
        xhr.send(formData);
      });

      setUploads((prev) =>
        prev.map((u) =>
          u.id === tempId ? { ...u, progress: 100, status: "done", media: result } : u,
        ),
      );
      setMediaItems((prev) => [...prev, result]);
      return result;
    } catch (err) {
      const message = err instanceof Error ? err.message : "Upload failed";
      setUploads((prev) =>
        prev.map((u) =>
          u.id === tempId ? { ...u, status: "error", error: message } : u,
        ),
      );
      return null;
    }
  }, []);

  const addFiles = useCallback(
    async (files: File[]) => {
      for (const file of files) {
        const error = validateFile(file);
        if (error) {
          return error;
        }
      }

      const currentMediaType = mediaItems[0]?.type;
      for (const file of files) {
        const config = ALLOWED_TYPES[file.type];
        if (!config) continue;

        if (currentMediaType === "video" || currentMediaType === "gif") {
          return "동영상/GIF는 다른 미디어와 함께 첨부할 수 없습니다.";
        }
        if (config.type === "video" || config.type === "gif") {
          if (mediaItems.length > 0) {
            return "동영상/GIF는 다른 미디어와 함께 첨부할 수 없습니다.";
          }
        }
        if (config.type === "image" && mediaItems.length >= 4) {
          return "이미지는 최대 4장까지 첨부할 수 있습니다.";
        }
      }

      await Promise.all(files.map(uploadFile));
      return null;
    },
    [mediaItems, validateFile, uploadFile],
  );

  const removeMedia = useCallback((mediaId: string) => {
    setMediaItems((prev) => prev.filter((m) => m.id !== mediaId));
    setUploads((prev) => prev.filter((u) => u.media?.id !== mediaId));
  }, []);

  const reset = useCallback(() => {
    setUploads([]);
    setMediaItems([]);
  }, []);

  return {
    uploads,
    mediaItems,
    isUploading,
    addFiles,
    removeMedia,
    reset,
  };
}
