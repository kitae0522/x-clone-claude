import { useState, useRef, useEffect } from "react";
import { Image, MapPin, BarChart3 } from "lucide-react";
import { useCreateReply } from "@/hooks/useReplies";
import { useMediaUpload } from "@/hooks/useMediaUpload";
import { useGeolocation } from "@/hooks/useGeolocation";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import MediaPreview from "@/components/MediaPreview";
import PollCreator from "@/components/PollCreator";

const MAX_LENGTH = 500;

interface ReplyFormProps {
  postId: string;
  parentPostId?: string;
}

export default function ReplyForm({ postId, parentPostId }: ReplyFormProps) {
  const [content, setContent] = useState("");
  const { mutate, isPending } = useCreateReply(postId, parentPostId);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const { uploads, mediaItems, isUploading, addFiles, removeMedia, reset } =
    useMediaUpload();

  const {
    location,
    isLoading: isLocationLoading,
    error: locationError,
    requestLocation,
    clearLocation,
  } = useGeolocation();

  const [showPoll, setShowPoll] = useState(false);
  const [pollOptions, setPollOptions] = useState(["", ""]);
  const [pollDuration, setPollDuration] = useState(1440);

  const remaining = MAX_LENGTH - [...content].length;

  useEffect(() => {
    if (locationError) {
      toast.error(locationError);
    }
  }, [locationError]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (remaining < 0 || isPending || isUploading) return;
    if (content.trim().length === 0 && mediaItems.length === 0) return;

    if (showPoll) {
      const filledOptions = pollOptions.filter((o) => o.trim().length > 0);
      if (filledOptions.length < 2) {
        toast.error("최소 2개의 선택지를 입력해주세요.");
        return;
      }
    }

    mutate(
      {
        content,
        mediaIds:
          mediaItems.length > 0 ? mediaItems.map((m) => m.id) : undefined,
        location: location
          ? {
              latitude: location.latitude,
              longitude: location.longitude,
              name: location.name,
            }
          : undefined,
        poll: showPoll
          ? {
              options: pollOptions.filter((o) => o.trim().length > 0),
              durationMinutes: pollDuration,
            }
          : undefined,
      },
      {
        onSuccess: () => {
          setContent("");
          reset();
          clearLocation();
          setShowPoll(false);
          setPollOptions(["", ""]);
          setPollDuration(1440);
          toast.success("댓글이 작성되었습니다.");
        },
        onError: (err) => {
          toast.error("댓글 작성에 실패했습니다.", {
            description: err.message,
          });
        },
      },
    );
  }

  async function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? []);
    if (files.length === 0) return;
    const error = await addFiles(files);
    if (error) toast.error(error);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  function handleMediaButtonClick() {
    if (showPoll) {
      toast.error("투표와 미디어는 동시에 첨부할 수 없습니다.");
      return;
    }
    fileInputRef.current?.click();
  }

  function handlePollToggle() {
    if (mediaItems.length > 0) {
      toast.error("미디어와 투표는 동시에 첨부할 수 없습니다.");
      return;
    }
    setShowPoll(!showPoll);
    if (!showPoll) {
      setPollOptions(["", ""]);
      setPollDuration(1440);
    }
  }

  function handleLocationClick() {
    if (location) {
      clearLocation();
    } else {
      requestLocation();
    }
  }

  return (
    <form className="border-b border-border p-4" onSubmit={handleSubmit}>
      <Textarea
        className="w-full resize-none border-none bg-transparent py-2 text-[15px] shadow-none focus-visible:ring-0 placeholder:text-muted-foreground"
        placeholder="Post your reply"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        maxLength={MAX_LENGTH}
        rows={2}
      />

      {/* Media preview */}
      <MediaPreview
        uploads={uploads}
        mediaItems={mediaItems}
        onRemove={removeMedia}
      />

      {/* Poll creator */}
      {showPoll && (
        <PollCreator
          options={pollOptions}
          durationMinutes={pollDuration}
          onOptionsChange={setPollOptions}
          onDurationChange={setPollDuration}
          onRemove={() => setShowPoll(false)}
        />
      )}

      {/* Location tag */}
      {location && (
        <div className="mt-2 flex items-center gap-1.5 text-[13px] text-primary">
          <MapPin size={14} />
          <span>{location.name}</span>
          <button
            type="button"
            onClick={clearLocation}
            className="cursor-pointer rounded-full border-none bg-transparent p-0.5 text-muted-foreground transition-colors hover:text-foreground"
          >
            x
          </button>
        </div>
      )}

      <div className="flex items-center justify-between border-t border-border pt-2">
        <div className="flex items-center gap-1">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/webp,image/gif,video/mp4,video/webm"
            multiple
            onChange={handleFileSelect}
            className="hidden"
          />
          <button
            type="button"
            onClick={handleMediaButtonClick}
            className="cursor-pointer rounded-full border-none bg-transparent p-1.5 text-primary/70 transition-colors hover:bg-primary/10 hover:text-primary"
            title="미디어 첨부"
          >
            <Image size={16} />
          </button>
          <button
            type="button"
            onClick={handleLocationClick}
            disabled={isLocationLoading}
            className={`cursor-pointer rounded-full border-none bg-transparent p-1.5 transition-colors hover:bg-primary/10 ${
              location ? "text-primary" : "text-primary/70 hover:text-primary"
            } disabled:opacity-50`}
            title="위치 추가"
          >
            <MapPin size={16} />
          </button>
          <button
            type="button"
            onClick={handlePollToggle}
            className={`cursor-pointer rounded-full border-none bg-transparent p-1.5 transition-colors hover:bg-primary/10 ${
              showPoll ? "text-primary" : "text-primary/70 hover:text-primary"
            }`}
            title="투표 만들기"
          >
            <BarChart3 size={16} />
          </button>
        </div>

        <div className="flex items-center gap-3">
          <span
            className={cn(
              "text-sm text-muted-foreground",
              remaining < 0 && "text-destructive",
              remaining >= 0 && remaining <= 20 && "text-warning",
            )}
          >
            {remaining}
          </span>
          <Button
            type="submit"
            size="sm"
            className="rounded-full"
            disabled={
              remaining < 0 ||
              (content.trim().length === 0 && mediaItems.length === 0) ||
              isPending ||
              isUploading
            }
          >
            {isPending ? "Replying..." : "Reply"}
          </Button>
        </div>
      </div>
    </form>
  );
}
