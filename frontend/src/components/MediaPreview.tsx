import { X } from "lucide-react";
import type { MediaItem } from "@/types/api";

interface UploadItem {
  id: string;
  progress: number;
  status: "uploading" | "processing" | "done" | "error";
  media?: MediaItem;
  error?: string;
}

interface MediaPreviewProps {
  uploads: UploadItem[];
  mediaItems: MediaItem[];
  onRemove: (mediaId: string) => void;
}

export default function MediaPreview({ uploads, mediaItems, onRemove }: MediaPreviewProps) {
  if (uploads.length === 0 && mediaItems.length === 0) return null;

  return (
    <div className="mt-2 flex flex-wrap gap-2">
      {uploads.map((upload) => {
        const media = upload.media ?? mediaItems.find((m) => m.id === upload.id);

        return (
          <div
            key={upload.id}
            className="relative h-24 w-24 overflow-hidden rounded-lg border border-border bg-muted"
          >
            {(upload.status === "uploading" || upload.status === "processing") && (
              <div className="flex h-full items-center justify-center">
                <div className="relative h-10 w-10">
                  <svg className="-rotate-90" viewBox="0 0 36 36">
                    <circle
                      cx="18"
                      cy="18"
                      r="14"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="3"
                      className="text-border"
                    />
                    <circle
                      cx="18"
                      cy="18"
                      r="14"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="3"
                      strokeDasharray={88}
                      strokeDashoffset={upload.status === "processing" ? 0 : 88 - (88 * upload.progress) / 100}
                      strokeLinecap="round"
                      className="text-primary transition-all duration-200"
                    />
                  </svg>
                  <span className="absolute inset-0 flex items-center justify-center text-[10px] font-medium text-foreground">
                    {upload.status === "processing" ? "변환중" : `${upload.progress}%`}
                  </span>
                </div>
              </div>
            )}

            {upload.status === "error" && (
              <div className="flex h-full items-center justify-center p-2">
                <span className="text-center text-[10px] text-destructive">
                  {upload.error ?? "실패"}
                </span>
              </div>
            )}

            {upload.status === "done" && media && (
              <>
                {media.type === "video" ? (
                  <video
                    src={media.url}
                    className="h-full w-full object-cover"
                    muted
                  />
                ) : (
                  <img
                    src={media.url}
                    alt=""
                    className="h-full w-full object-cover"
                  />
                )}
                <button
                  type="button"
                  onClick={() => onRemove(media.id)}
                  className="absolute right-1 top-1 cursor-pointer rounded-full border-none bg-black/60 p-1 text-white transition-colors hover:bg-black/80"
                >
                  <X size={12} />
                </button>
                {media.type === "gif" && (
                  <span className="absolute bottom-1 left-1 rounded bg-black/70 px-1 py-0.5 text-[9px] font-bold text-white">
                    GIF
                  </span>
                )}
              </>
            )}
          </div>
        );
      })}
    </div>
  );
}
