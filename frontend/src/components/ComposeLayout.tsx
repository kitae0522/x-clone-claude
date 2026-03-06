import { useState, useRef, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, Image, MapPin, BarChart3 } from "lucide-react";
import { useAuth } from "@/hooks/useAuthContext";
import { useMediaUpload } from "@/hooks/useMediaUpload";
import { useGeolocation } from "@/hooks/useGeolocation";
import { toast } from "sonner";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import UserAvatar from "@/components/UserAvatar";
import MarkdownRenderer from "@/components/MarkdownRenderer";
import MediaPreview from "@/components/MediaPreview";
import PollCreator from "@/components/PollCreator";
import VisibilitySelector from "@/components/VisibilitySelector";
import type { MediaItem, LocationData } from "@/types/api";

const MAX_LENGTH = 500;
const WARN_THRESHOLD = 20;
const CIRCLE_RADIUS = 10;
const CIRCLE_CIRCUMFERENCE = 2 * Math.PI * CIRCLE_RADIUS;

export interface ComposeSubmitData {
  content: string;
  visibility: "public" | "follower" | "private";
  mediaIds: string[];
  location: { latitude: number; longitude: number; name: string } | null;
  poll: { options: string[]; durationMinutes: number } | null;
}

interface ComposeLayoutProps {
  onSubmit: (data: ComposeSubmitData) => void;
  isPending: boolean;
  submitLabel: string;
  pendingLabel: string;
  placeholder?: string;
  showVisibility?: boolean;
  initialContent?: string;
  initialVisibility?: "public" | "follower" | "private";
  initialLocation?: LocationData | null;
  initialMediaItems?: MediaItem[];
  initialPollOptions?: string[];
  initialPollDuration?: number;
  children?: React.ReactNode;
}

export default function ComposeLayout({
  onSubmit,
  isPending,
  submitLabel,
  pendingLabel,
  placeholder = "무슨 일이 일어나고 있나요?",
  showVisibility = true,
  initialContent = "",
  initialVisibility = "public",
  initialLocation = null,
  initialMediaItems = [],
  initialPollOptions,
  initialPollDuration = 1440,
  children,
}: ComposeLayoutProps) {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [content, setContent] = useState(initialContent);
  const [visibility, setVisibility] = useState(initialVisibility);
  const [showPreview, setShowPreview] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const { uploads, mediaItems, isUploading, addFiles, removeMedia } =
    useMediaUpload(initialMediaItems);

  const {
    location,
    isLoading: isLocationLoading,
    error: locationError,
    requestLocation,
    clearLocation,
  } = useGeolocation(initialLocation);

  const [showPoll, setShowPoll] = useState(!!initialPollOptions);
  const [pollOptions, setPollOptions] = useState(
    initialPollOptions ?? ["", ""],
  );
  const [pollDuration, setPollDuration] = useState(initialPollDuration);

  const charCount = [...content].length;
  const remaining = MAX_LENGTH - charCount;
  const progress = Math.min(charCount / MAX_LENGTH, 1);
  const strokeDashoffset = CIRCLE_CIRCUMFERENCE * (1 - progress);

  const circleColor =
    remaining < 0
      ? "text-destructive"
      : remaining <= WARN_THRESHOLD
        ? "text-warning"
        : "text-primary";

  const hasMarkdown = /[*_`~>#\-[\]()]/.test(content);
  const hasContent = content.trim().length > 0 || mediaItems.length > 0;
  const canSubmit = hasContent && remaining >= 0 && !isPending && !isUploading;

  useEffect(() => {
    textareaRef.current?.focus();
  }, []);

  useEffect(() => {
    if (locationError) {
      toast.error(locationError);
    }
  }, [locationError]);

  function handleSubmit() {
    if (!canSubmit) return;

    if (showPoll) {
      const filledOptions = pollOptions.filter((o) => o.trim().length > 0);
      if (filledOptions.length < 2) {
        toast.error("최소 2개의 선택지를 입력해주세요.");
        return;
      }
    }

    onSubmit({
      content,
      visibility,
      mediaIds: mediaItems.map((m) => m.id),
      location: location
        ? {
            latitude: location.latitude,
            longitude: location.longitude,
            name: location.name,
          }
        : null,
      poll: showPoll
        ? {
            options: pollOptions.filter((o) => o.trim().length > 0),
            durationMinutes: pollDuration,
          }
        : null,
    });
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
      setPollOptions(initialPollOptions ?? ["", ""]);
      setPollDuration(initialPollDuration);
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
    <div className="min-h-dvh">
      {/* Header */}
      <header className="sticky top-0 z-10 flex items-center justify-between border-b border-border bg-background/65 px-4 py-2 backdrop-blur-xl">
        <button
          onClick={() => navigate(-1)}
          className="cursor-pointer rounded-full border-none bg-transparent p-2 transition-colors hover:bg-foreground/10"
        >
          <ArrowLeft size={20} />
        </button>
        <Button
          onClick={handleSubmit}
          className="rounded-full px-5"
          size="sm"
          disabled={!canSubmit}
        >
          {isPending ? pendingLabel : submitLabel}
        </Button>
      </header>

      {/* Context slot (reply context, etc.) */}
      {children}

      {/* Compose area */}
      <div className={`flex gap-3 px-4 ${!children ? "pt-3" : ""}`}>
        <UserAvatar
          profileImageUrl={user?.profileImageUrl}
          displayName={user?.displayName}
          size="md"
          className="mt-1 shrink-0"
        />
        <div className="flex-1">
          {/* Visibility selector */}
          {showVisibility && (
            <VisibilitySelector value={visibility} onChange={setVisibility} />
          )}

          <Textarea
            ref={textareaRef}
            className="w-full resize-none border-none bg-transparent py-2 text-xl shadow-none focus-visible:ring-0 placeholder:text-muted-foreground"
            placeholder={placeholder}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={5}
          />

          {/* Live markdown preview */}
          {showPreview && content.trim().length > 0 && (
            <div className="mt-2 rounded-lg border border-border bg-muted/30 p-3">
              <MarkdownRenderer content={content} className="text-[15px]" />
            </div>
          )}

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

          {/* Toolbar */}
          <div className="mt-2 flex items-center justify-between border-t border-border pt-2">
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
                className="cursor-pointer rounded-full border-none bg-transparent p-2 text-primary/70 transition-colors hover:bg-primary/10 hover:text-primary"
                title="미디어 첨부"
              >
                <Image size={18} />
              </button>
              <button
                type="button"
                onClick={handleLocationClick}
                disabled={isLocationLoading}
                className={`cursor-pointer rounded-full border-none bg-transparent p-2 transition-colors hover:bg-primary/10 ${
                  location
                    ? "text-primary"
                    : "text-primary/70 hover:text-primary"
                } disabled:opacity-50`}
                title="위치 추가"
              >
                <MapPin size={18} />
              </button>
              <button
                type="button"
                onClick={handlePollToggle}
                className={`cursor-pointer rounded-full border-none bg-transparent p-2 transition-colors hover:bg-primary/10 ${
                  showPoll
                    ? "text-primary"
                    : "text-primary/70 hover:text-primary"
                }`}
                title="투표 만들기"
              >
                <BarChart3 size={18} />
              </button>

              {hasMarkdown && (
                <button
                  type="button"
                  onClick={() => setShowPreview(!showPreview)}
                  className={`cursor-pointer rounded-full border-none px-3 py-1 text-xs font-medium transition-colors ${
                    showPreview
                      ? "bg-primary text-primary-foreground"
                      : "bg-muted text-muted-foreground hover:bg-muted/80"
                  }`}
                >
                  {showPreview ? "미리보기 닫기" : "Md"}
                </button>
              )}
            </div>

            {charCount > 0 && (
              <div className="flex items-center gap-1.5">
                <svg
                  className="h-[26px] w-[26px] -rotate-90"
                  viewBox="0 0 24 24"
                >
                  <circle
                    cx="12"
                    cy="12"
                    r={CIRCLE_RADIUS}
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    className="text-border"
                  />
                  <circle
                    cx="12"
                    cy="12"
                    r={CIRCLE_RADIUS}
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeDasharray={CIRCLE_CIRCUMFERENCE}
                    strokeDashoffset={strokeDashoffset}
                    strokeLinecap="round"
                    className={`transition-all duration-200 ${circleColor}`}
                  />
                </svg>
                {remaining <= WARN_THRESHOLD && (
                  <span
                    className={`text-[13px] font-medium ${remaining < 0 ? "text-destructive" : "text-warning"}`}
                  >
                    {remaining}
                  </span>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
