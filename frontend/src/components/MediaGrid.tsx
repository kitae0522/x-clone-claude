import { useState } from "react";
import type { MediaItem } from "@/types/api";
import { cn } from "@/lib/utils";

interface MediaGridProps {
  media: MediaItem[];
}

export default function MediaGrid({ media }: MediaGridProps) {
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null);

  if (!media || media.length === 0) return null;

  const count = media.length;

  function handleClick(e: React.MouseEvent, index: number) {
    e.stopPropagation();
    setLightboxIndex(index);
  }

  return (
    <>
      <div
        className={cn(
          "mt-3 overflow-hidden rounded-2xl border border-border",
          count === 1 && "grid grid-cols-1",
          count === 2 && "grid grid-cols-2 gap-0.5",
          count === 3 && "grid grid-cols-2 gap-0.5",
          count === 4 && "grid grid-cols-2 gap-0.5",
        )}
        onClick={(e) => e.stopPropagation()}
      >
        {media.map((item, i) => (
          <div
            key={item.id}
            className={cn(
              "relative cursor-pointer overflow-hidden bg-muted",
              count === 1 && "max-h-[512px]",
              count === 2 && "aspect-[4/5]",
              count === 3 && i === 0 && "row-span-2 aspect-auto h-full",
              count === 3 && i > 0 && "aspect-square",
              count === 4 && "aspect-square",
            )}
            onClick={(e) => handleClick(e, i)}
          >
            {item.type === "video" ? (
              <video
                src={item.url}
                controls
                className="h-full w-full object-cover"
                onClick={(e) => e.stopPropagation()}
              />
            ) : (
              <img
                src={item.url}
                alt=""
                className="h-full w-full object-cover transition-opacity hover:opacity-90"
                loading="lazy"
              />
            )}
            {item.type === "gif" && (
              <span className="absolute bottom-2 left-2 rounded bg-black/70 px-1.5 py-0.5 text-[11px] font-bold text-white">
                GIF
              </span>
            )}
          </div>
        ))}
      </div>

      {/* Lightbox */}
      {lightboxIndex !== null && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/90"
          onClick={() => setLightboxIndex(null)}
        >
          <button
            className="absolute right-4 top-4 cursor-pointer rounded-full bg-white/10 p-2 text-white transition-colors hover:bg-white/20"
            onClick={() => setLightboxIndex(null)}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>

          {lightboxIndex > 0 && (
            <button
              className="absolute left-4 cursor-pointer rounded-full bg-white/10 p-2 text-white transition-colors hover:bg-white/20"
              onClick={(e) => { e.stopPropagation(); setLightboxIndex(lightboxIndex - 1); }}
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
            </button>
          )}

          {lightboxIndex < media.length - 1 && (
            <button
              className="absolute right-4 cursor-pointer rounded-full bg-white/10 p-2 text-white transition-colors hover:bg-white/20"
              onClick={(e) => { e.stopPropagation(); setLightboxIndex(lightboxIndex + 1); }}
              style={{ top: "50%", transform: "translateY(-50%)" }}
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
            </button>
          )}

          <div className="max-h-[90vh] max-w-[90vw]" onClick={(e) => e.stopPropagation()}>
            {media[lightboxIndex].type === "video" ? (
              <video
                src={media[lightboxIndex].url}
                controls
                autoPlay
                className="max-h-[90vh] max-w-[90vw]"
              />
            ) : (
              <img
                src={media[lightboxIndex].url}
                alt=""
                className="max-h-[90vh] max-w-[90vw] object-contain"
              />
            )}
          </div>
        </div>
      )}
    </>
  );
}
